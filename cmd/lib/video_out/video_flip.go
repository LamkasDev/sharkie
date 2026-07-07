package video_out

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000C6C0
// __int64 __fastcall sceVideoOutAddFlipEvent(unsigned int, int, __int64, double)
func libSceVideoOut_sceVideoOutAddFlipEvent(equeueHandle, rawHandle, userData uintptr) uintptr {
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutAddFlipEvent"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	event := KernelEvent{
		Id:       uint64(handle.Id),
		Filter:   EVFILT_VBLANK,
		Flags:    EV_ADD,
		UserData: userData,
	}
	result := kernel.Kevent(equeueHandle, uintptr(unsafe.Pointer(&event)), 1, 0, 0, 0)

	logger.Printf("%-132s %s added flip event to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutAddFlipEvent"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return result
}

func SceVideoOutSubmitEopFlip(rawHandle, bufferIndex, flipMode, flipArg, eopSignalCtx uintptr) uintptr {
	return libSceVideoOut_sceVideoOutSubmitEopFlip(rawHandle, bufferIndex, flipMode, flipArg, eopSignalCtx)
}

// 0x000000000000B950
// __int64 __fastcall sceVideoOutSubmitEopFlip(int a1, unsigned int a2, unsigned int a3, __int64 a4, __int64 a5)
func libSceVideoOut_sceVideoOutSubmitEopFlip(rawHandle, bufferIndex, flipMode, flipArg, eopSignalCtx uintptr) uintptr {
	if int(bufferIndex) >= VideoOutMaxBuffers || bufferIndex == 0xFFFFFFFF {
		logger.Printf("%-132s %s failed due to invalid buffer index.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	buffer := &handle.Buffers[bufferIndex]
	if !buffer.Registered {
		logger.Printf("%-132s %s failed due to %s's buffer slot %s not being registered.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
			color.Yellow.Sprintf("0x%X", handle.Id),
			color.Yellow.Sprintf("0x%X", bufferIndex),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}

	// Ask GPU to present new buffer.
	handle.CurrentBuffer = uint32(bufferIndex)
	GlobalLiverpool.Flip(buffer.GpuAddress, uint64(flipArg))

	// TODO: actually sync it with the ticker (too lazy).
	time.Sleep(16666 * time.Microsecond)

	// Simulate EOP completion.
	if handle.LabelBufferAddress != 0 {
		labelSlot := (*uint64)(unsafe.Pointer(handle.LabelBufferAddress + bufferIndex*8))
		*labelSlot = 1
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s submitted %s's EOP flip (bufferIndex=%s, flipMode=%s, flipArg=%s, eopSignalCtx=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
			color.Yellow.Sprintf("0x%X", handle.Id),
			color.Yellow.Sprintf("0x%X", bufferIndex),
			color.Yellow.Sprintf("0x%X", flipMode),
			color.Yellow.Sprintf("0x%X", flipArg),
			color.Yellow.Sprintf("0x%X", eopSignalCtx),
		)
	}
	return 0
}
