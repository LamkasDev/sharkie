package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/gookit/color"
)

// 0x000000000000B620
// __int64 __fastcall sceVideoOutRegisterBuffers(int, unsigned int, __int64, unsigned int, __int64)
func libSceVideoOut_sceVideoOutRegisterBuffers(rawHandle, startIndex, addressesPtr, bufferNum, attrPtr uintptr) uintptr {
	if addressesPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid adresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	if attrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	end := int(startIndex) + int(bufferNum)
	if int(startIndex) < 0 || end > VideoOutMaxBuffers {
		logger.Printf("%-132s %s failed due to too exceeding maximum number of buffers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}

	attribute := (*VideoOutBufferAttribute)(unsafe.Pointer(attrPtr))
	addresses := unsafe.Slice((*uintptr)(unsafe.Pointer(addressesPtr)), bufferNum)
	handle.Attributes[0] = *attribute
	for i := range bufferNum {
		slot := startIndex + i
		address := addresses[i]
		handle.Buffers[slot] = VideoOutBuffer{
			GpuAddress:     address,
			AttributeIndex: 0,
			Registered:     true,
		}
		GlobalLiverpool.RegisterDisplaySurface(address, attribute, 0)

		logger.Printf("%-132s %s registered %s's buffer slot %s (address=%s, pixf=%s, tile=%s, aspr=%s, %sx%s, pitch=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
			color.Yellow.Sprintf("0x%X", handle.Id),
			color.Yellow.Sprintf("%d", slot),
			color.Yellow.Sprintf("0x%X", address),
			color.Yellow.Sprintf("0x%X", attribute.PixelFormat),
			color.Yellow.Sprintf("0x%X", attribute.TilingMode),
			color.Yellow.Sprintf("0x%X", attribute.AspectRatio),
			color.Yellow.Sprintf("%d", attribute.Width),
			color.Yellow.Sprintf("%d", attribute.Height),
			color.Yellow.Sprintf("%d", attribute.PitchInPixel),
		)
	}

	return 0
}

// 0x000000000000B240
// __int64 __fastcall sceVideoOutRegisterBufferAttribute(int, unsigned int, __int64)
func libSceVideoOut_sceVideoOutRegisterBufferAttribute(rawHandle, attributeIndex, attrPtr uintptr) uintptr {
	if attrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBufferAttribute"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBufferAttribute"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	if int(attributeIndex) >= len(handle.Attributes) {
		logger.Printf("%-132s %s failed due to invalid attribute index (%s >= %s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBufferAttribute"),
			color.Yellow.Sprintf("%d", attributeIndex),
			color.Yellow.Sprintf("%d", len(handle.Attributes)),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}

	attribute := (*VideoOutBufferAttribute)(unsafe.Pointer(attrPtr))
	handle.Attributes[attributeIndex] = *attribute

	logger.Printf("%-132s %s registered %s's buffer attribute %s in %s (pixf=%s, tile=%s, aspr=%s, %sx%s, pitch=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Yellow.Sprintf("%d", attributeIndex),
		color.Yellow.Sprintf("0x%X", attribute.PixelFormat),
		color.Yellow.Sprintf("0x%X", attribute.TilingMode),
		color.Yellow.Sprintf("0x%X", attribute.AspectRatio),
		color.Yellow.Sprintf("%d", attribute.Width),
		color.Yellow.Sprintf("%d", attribute.Height),
		color.Yellow.Sprintf("%d", attribute.PitchInPixel),
	)
	return 0
}

// 0x0000000000002860
// __int64 __fastcall sceVideoOutSetBufferAttribute(_DWORD *_RDI, int, int, int, int, int, __m128 _XMM0, unsigned int)
func libSceVideoOut_sceVideoOutSetBufferAttribute(attrPtr, pixelFormat, tilingMode, aspectRatio, width, height, pitchInPixel uintptr) uintptr {
	if attrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSetBufferAttribute"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}

	attribute := (*VideoOutBufferAttribute)(unsafe.Pointer(attrPtr))
	attribute.PixelFormat = VideoOutPixelFormat(pixelFormat)
	attribute.TilingMode = VideoOutTilingMode(tilingMode)
	attribute.AspectRatio = VideoOutAspectRatio(aspectRatio)
	attribute.Width = uint32(width)
	attribute.Height = uint32(height)
	attribute.PitchInPixel = uint32(pitchInPixel)

	logger.Printf("%-132s %s set buffer attribute %s (pixf=%s, tile=%s, aspr=%s, %sx%s, pitch=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSetBufferAttribute"),
		color.Yellow.Sprintf("0x%X", attrPtr),
		color.Yellow.Sprintf("0x%X", attribute.PixelFormat),
		color.Yellow.Sprintf("0x%X", attribute.TilingMode),
		color.Yellow.Sprintf("0x%X", attribute.AspectRatio),
		color.Yellow.Sprintf("%d", attribute.Width),
		color.Yellow.Sprintf("%d", attribute.Height),
		color.Yellow.Sprintf("%d", attribute.PitchInPixel),
	)
	return 0
}
