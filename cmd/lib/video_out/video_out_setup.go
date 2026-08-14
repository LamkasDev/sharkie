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

	var counter uint64
	for range ticker.C {
		// Construct filter data components.
		timeBits := uint64(time.Now().UnixNano() & 0xFFF)
		if counter != 0xF {
			counter++
		}
		counterBits := counter << 12

		// Send v-blank events.
		vblankData := timeBits | counterBits | ((handle.VblankStatus.Count << 16) & 0xFFFFFFFFFFFF0000)
		for _, listener := range handle.VblankEvents {
			if equeue := GetEqueue(listener.EqueueHandle); equeue != nil {
				kevent := KernelEvent{
					Id:         VideoOutInternalEventIdVblank,
					Filter:     KernelEventFilterVideoOut,
					FilterData: vblankData,
					UserData:   listener.UserData,
				}
				select {
				case equeue.Events <- kevent:
				default:
				}
			}
		}

		// Update vblank status.
		handle.VblankStatus.Count++
		handle.VblankStatus.ProcessTime = uint64(kernel.SceKernelGetProcessTime())
		handle.VblankStatus.Tsc = uint64(ReadTsc())

		// Check if we need to flip.
		if handle.VblankStatus.Count%(uint64(handle.FlipRate)+1) != 0 {
			continue
		}

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

		// Submit flip to Vulkan.
		gpu.GlobalLiverpool.OnFlip(handle.CurrentFlip)
		oldLabelAddress := handle.LabelBufferAddress + uintptr(handle.CurrentFlip.BufferIndex)*8
		*(*uint64)(unsafe.Pointer(oldLabelAddress)) = 0

		// Update flip status.
		handle.FlipStatus.Count++
		handle.FlipStatus.ProcessTime = uint64(kernel.SceKernelGetProcessTime())
		handle.FlipStatus.Tsc = uint64(ReadTsc())
		handle.FlipStatus.FlipArg = handle.CurrentFlip.FlipArg
		handle.FlipStatus.CurrentBuffer = handle.CurrentFlip.BufferIndex
		handle.FlipStatus.GcQueueNumber--
		handle.FlipStatus.FlipPendingNumber--

		// Send flip events.
		flipData := timeBits | counterBits | ((uint64(handle.CurrentFlip.FlipArg) << 16) & 0xFFFFFFFFFFFF0000)
		for _, listener := range handle.FlipEvents {
			if equeue := GetEqueue(listener.EqueueHandle); equeue != nil {
				kevent := KernelEvent{
					Id:         VideoOutInternalEventIdFlip,
					Filter:     KernelEventFilterVideoOut,
					FilterData: flipData,
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
func libSceVideoOut_sceVideoOutGetResolutionStatus(rawHandle uintptr, resolutionStatus *VideoOutResolutionStatus) uintptr {
	if resolutionStatus == nil {
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

// 0x00000000000043B0
// __int64 __fastcall sceVideoOutGetDeviceCapabilityInfo_(int, __int64, __int64)
func libSceVideoOut_sceVideoOutGetDeviceCapabilityInfo_(rawHandle uintptr, capabilityInfo *VideoOutCapabilityInfo) uintptr {
	if capabilityInfo == nil {
		logger.Printf("%-132s %s failed due to invalid capability info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetDeviceCapabilityInfo_"),
		)
		return 0x80290002
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetDeviceCapabilityInfo_"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	*capabilityInfo = handle.Capability

	if logger.LogGraphics {
		logger.Printf("%-132s %s returned %s's capability info.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetDeviceCapabilityInfo_"),
			color.Yellow.Sprintf("0x%X", handle.Id),
		)
	}
	return 0
}
