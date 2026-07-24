package video_out

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
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
	handle := NewVideoOutHandle(GlobalDisplayCoreEngine.NextHandle)
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
		if handle.CurrentFlip == nil {
			select {
			case nextFlip := <-handle.NextFlip:
				handle.CurrentFlip = nextFlip
			default:
			}
		}
		if handle.CurrentFlip == nil {
			continue
		}

		// Submit flip to Vulkan if ready.
		labelSlot := (*uint64)(unsafe.Pointer(handle.LabelBufferAddress + uintptr(handle.CurrentFlip.BufferIndex)*8))
		if *labelSlot == 0 {
			continue
		}
		gpu.GlobalLiverpool.OnFlip(handle.CurrentFlip)
		oldLabelAddress := handle.LabelBufferAddress + uintptr(handle.CurrentFlip.BufferIndex)*8
		*(*uint64)(unsafe.Pointer(oldLabelAddress)) = 0

		// Update flip status.
		handle.FlipStatus.Count++
		handle.FlipStatus.ProcessTime = uint64(kernel.SceKernelGetProcessTime())
		handle.FlipStatus.Tsc = uint64(kernel.SceKernelReadTsc())
		handle.FlipStatus.FlipArg = handle.CurrentFlip.FlipArg
		handle.FlipStatus.CurrentBuffer = handle.CurrentFlip.BufferIndex
		handle.FlipStatus.GcQueueNumber--
		handle.FlipStatus.FlipPendingNumber--

		// Construct filter data components.
		timeBits := uint64(time.Now().UnixNano() & 0xFFF)
		if counter != 0xF {
			counter++
		}
		counterBits := counter << 12

		// Send flip events.
		flipData := timeBits | counterBits | ((uint64(handle.CurrentFlip.FlipArg) << 16) & 0xFFFFFFFFFFFF0000)
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

		// Reset flip.
		handle.CurrentFlip = nil
	}
}

// 0x000000000000BE60
// __int64 __fastcall sceVideoOutGetResolutionStatus(int, __int64)
func libSceVideoOut_sceVideoOutGetResolutionStatus(rawHandle, resolutionStatusPtr uintptr) uintptr {
	if resolutionStatusPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid resolution status pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetResolutionStatus"),
		)
		return 0x80290002
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetResolutionStatus"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	resolutionStatus := (*VideoOutResolutionStatus)(unsafe.Pointer(resolutionStatusPtr))
	*resolutionStatus = handle.Resolution

	if logger.LogGraphics {
		logger.Printf("%-132s %s returned %s's resolution status.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetResolutionStatus"),
			color.Yellow.Sprintf("0x%X", handle.Id),
		)
	}
	return 0
}
