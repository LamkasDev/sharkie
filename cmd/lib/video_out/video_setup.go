package video_out

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000AAD0
// __int64 __fastcall sceVideoOutOpen(unsigned int, unsigned int, unsigned int, _DWORD *, __m128 _XMM0)
func libSceVideoOut_sceVideoOutOpen() uintptr {
	handle := &VideoOutHandle{
		Id:                 GlobalDisplayCoreEngine.NextHandle,
		LabelBufferAddress: GlobalGoAllocator.Malloc(uintptr(VideoOutMaxBuffers) * 8),
		NextFlip:           make(chan *VideoOutFlip, VideoOutMaxBuffers),
	}
	GlobalDisplayCoreEngine.Handles[handle.Id] = handle
	GlobalDisplayCoreEngine.NextHandle++

	go pprof.Do(context.Background(), pprof.Labels("name", fmt.Sprintf("VblankTicker-%08x", handle.Id)), func(ctx context.Context) {
		displayVblankTicker(handle)
	})

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}

func displayVblankTicker(handle *VideoOutHandle) {
	ticker := time.NewTicker(16666 * time.Microsecond)
	defer ticker.Stop()

	var vblankCount, counter uint64
	for range ticker.C {
		vblankCount++

		// Pull flip requests.
		if handle.StagingFlip == nil {
			select {
			case nextFlip := <-handle.NextFlip:
				handle.StagingFlip = nextFlip
			default:
			}
		}

		// Submit flip to Vulkan if ready.
		if handle.StagingFlip != nil {
			labelSlot := (*uint64)(unsafe.Pointer(handle.LabelBufferAddress + uintptr(handle.StagingFlip.BufferIndex)*8))
			if *labelSlot == 1 {
				gpu.GlobalLiverpool.OnFlip(handle.StagingFlip)
				if handle.CurrentFlip != nil {
					oldLabelAddress := handle.LabelBufferAddress + uintptr(handle.CurrentFlip.BufferIndex)*8
					*(*uint64)(unsafe.Pointer(oldLabelAddress)) = 0
				}
				handle.CurrentFlip = handle.StagingFlip
				handle.StagingFlip = nil
			}
		}

		// Construct filter data components.
		timeBits := uint64(time.Now().UnixNano() & 0xFFF)
		if counter != 0xF {
			counter++
		}
		counterBits := counter << 12

		// Send flip events.
		var currentFlipArg uint64
		if handle.CurrentFlip != nil {
			currentFlipArg = handle.CurrentFlip.FlipArg
		}
		flipData := timeBits | counterBits | ((currentFlipArg << 16) & 0xFFFFFFFFFFFF0000)
		for _, listener := range handle.FlipEvents {
			if equeue := GetEqueue(listener.EqueueHandle); equeue != nil {
				kevent := KernelEvent{
					Id:         VideoOutInternalEventIdFlip,
					Filter:     EVFILT_VIDEO_OUT,
					FilterData: flipData,
					UserData:   listener.UserData,
				}
				select {
				case equeue.Events <- kevent:
				default:
				}
			}
		}

		// Send v-blank events.
		vblankData := timeBits | counterBits | ((vblankCount << 16) & 0xFFFFFFFFFFFF0000)
		for _, listener := range handle.VblankEvents {
			if equeue := GetEqueue(listener.EqueueHandle); equeue != nil {
				kevent := KernelEvent{
					Id:         VideoOutInternalEventIdVblank,
					Filter:     EVFILT_VIDEO_OUT,
					FilterData: vblankData,
					UserData:   listener.UserData,
				}
				select {
				case equeue.Events <- kevent:
				default:
				}
			}
		}
	}
}
