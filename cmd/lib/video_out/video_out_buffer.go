package video_out

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/gookit/color"
)

// 0x000000000000B620
// __int64 __fastcall sceVideoOutRegisterBuffers(int, unsigned int, __int64, unsigned int, __int64)
func libSceVideoOut_sceVideoOutRegisterBuffers(handleId uint32, startIndex, addressesPtr, bufferNum uintptr, attribute *VideoOutBufferAttribute) uintptr {
	if addressesPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid adresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	if attribute == nil {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBuffers"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[handleId]
	if !ok {
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

	addresses := unsafe.Slice((*uintptr)(unsafe.Pointer(addressesPtr)), bufferNum)
	handle.Attributes[0] = *attribute
	for i := range bufferNum {
		slot := startIndex + i
		address := structs.GetPhysicalGpuAddress(addresses[i])
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
func libSceVideoOut_sceVideoOutRegisterBufferAttribute(handleId uint32, attributeIndex uintptr, attribute *VideoOutBufferAttribute) uintptr {
	if attribute == nil {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutRegisterBufferAttribute"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[handleId]
	if !ok {
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

	handle.Attributes[attributeIndex] = *attribute

	logger.Printf("%-132s %s registered %s's buffer attribute %s (pixf=%s, tile=%s, aspr=%s, %sx%s, pitch=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutRegisterBufferAttribute"),
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
func libSceVideoOut_sceVideoOutSetBufferAttribute(attribute *VideoOutBufferAttribute, pixelFormat, tilingMode, aspectRatio uintptr, width, height, pitchInPixel uint32) uintptr {
	if attribute == nil {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSetBufferAttribute"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}

	attribute.PixelFormat = VideoOutPixelFormat(pixelFormat)
	attribute.TilingMode = VideoOutTilingMode(tilingMode)
	attribute.AspectRatio = VideoOutAspectRatio(aspectRatio)
	attribute.Width = width
	attribute.Height = height
	attribute.PitchInPixel = pitchInPixel

	logger.Printf("%-132s %s set buffer attribute %s (pixf=%s, tile=%s, aspr=%s, %sx%s, pitch=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSetBufferAttribute"),
		color.Yellow.Sprintf("0x%X", attribute),
		color.Yellow.Sprintf("0x%X", attribute.PixelFormat),
		color.Yellow.Sprintf("0x%X", attribute.TilingMode),
		color.Yellow.Sprintf("0x%X", attribute.AspectRatio),
		color.Yellow.Sprintf("%d", attribute.Width),
		color.Yellow.Sprintf("%d", attribute.Height),
		color.Yellow.Sprintf("%d", attribute.PitchInPixel),
	)
	return 0
}

func SceVideoOutGetBufferLabelAddress(handleId uint32, resultLabelBufferAddressPtr *uintptr) uintptr {
	return libSceVideoOut_sceVideoOutGetBufferLabelAddress(handleId, resultLabelBufferAddressPtr)
}

// 0x000000000000BB80
// __int64 __fastcall sceVideoOutGetBufferLabelAddress(int, _QWORD *)
func libSceVideoOut_sceVideoOutGetBufferLabelAddress(handleId uint32, resultLabelBufferAddressPtr *uintptr) uintptr {
	if resultLabelBufferAddressPtr == nil {
		logger.Printf("%-132s %s failed due to invalid result label buffer address pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[handleId]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	*resultLabelBufferAddressPtr = handle.LabelBufferAddress

	if logger.LogGraphics {
		logger.Printf("%-132s %s wrote %s's label buffer address %s to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
			color.Yellow.Sprintf("0x%X", handle.Id),
			color.Yellow.Sprintf("0x%X", handle.LabelBufferAddress),
			color.Yellow.Sprintf("0x%X", resultLabelBufferAddressPtr),
		)
	}
	return 0
}
