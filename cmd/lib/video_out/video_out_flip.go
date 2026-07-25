package video_out

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
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

	handle.FlipEvents = append(handle.FlipEvents, VideoOutEvent{
		EqueueHandle: equeueHandle,
		UserData:     userData,
	})

	logger.Printf("%-132s %s added flip event to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutAddFlipEvent"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return 0
}

// 0x000000000000BDE0
// __int64 __fastcall sceVideoOutSetFlipRate(int, unsigned int)
func libSceVideoOut_sceVideoOutSetFlipRate(rawHandle, flipRate uintptr) uintptr {
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSetFlipRate"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	handle.FlipRate = uint32(flipRate)

	logger.Printf("%-132s %s set %s's flip rate to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSetFlipRate"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Green.Sprintf("%d", flipRate),
	)
	return 0
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
	handle.SubmitFlip(&VideoOutFlip{
		BufferIndex: int32(bufferIndex),
		FlipArg:     int64(flipArg),
		GpuAddress:  buffer.GpuAddress,
	})

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

// 0x000000000000BA20
// __int64 __fastcall sceVideoOutGetFlipStatus(int, __int64)
func libSceVideoOut_sceVideoOutGetFlipStatus(rawHandle, flipStatusPtr uintptr) uintptr {
	if flipStatusPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid flip status pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetFlipStatus"),
		)
		return 0x80290002
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetFlipStatus"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	flipStatus := (*VideoOutFlipStatus)(unsafe.Pointer(flipStatusPtr))
	*flipStatus = handle.FlipStatus

	if false && logger.LogGraphics {
		logger.Printf("%-132s %s returned %s's flip status.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetFlipStatus"),
			color.Yellow.Sprintf("0x%X", handle.Id),
		)
	}
	return 0
}
