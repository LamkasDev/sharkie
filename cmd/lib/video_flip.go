package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/gookit/color"
)

// 0x000000000000C6C0
// __int64 __fastcall sceVideoOutAddFlipEvent(unsigned int, int, __int64, double)
func libSceVideoOut_sceVideoOutAddFlipEvent(equeueHandle, rawHandle, userData uintptr) uintptr {
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutAddFlipEvent"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	event := Kevent{
		Id:       uint64(handle.Id),
		Filter:   EVFILT_VBLANK,
		Flags:    EV_ADD,
		UserData: userData,
	}
	result := libKernel_kevent(equeueHandle, uintptr(unsafe.Pointer(&event)), 1, 0, 0, 0)

	logger.Printf("%-132s %s added flip event to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutAddFlipEvent"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return result
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
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
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

	// Simulate EOP completion.
	if handle.LabelBufferAddress != 0 {
		labelSlot := (*uint64)(unsafe.Pointer(handle.LabelBufferAddress + bufferIndex*8))
		*labelSlot = 1
	}

	logger.Printf("%-132s %s submitted %s's EOP flip (bufferIndex=%s, flipMode=%s, flipArg=%s, eopSignalCtx=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Yellow.Sprintf("0x%X", bufferIndex),
		color.Yellow.Sprintf("0x%X", flipMode),
		color.Yellow.Sprintf("0x%X", flipArg),
		color.Yellow.Sprintf("0x%X", eopSignalCtx),
	)
	return 0
}
