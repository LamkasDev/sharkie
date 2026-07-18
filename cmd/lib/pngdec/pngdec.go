package pngdec

import (
	"bytes"
	"image"
	"image/png"

	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pngdec"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPngDecStubs() {
	return
	elf.RegisterStub("libScePngDec", "scePngDecQueryMemorySize", libPngDec_scePngDecQueryMemorySize)
	elf.RegisterStub("libScePngDec", "scePngDecCreate", libPngDec_scePngDecCreate)
	elf.RegisterStub("libScePngDec", "scePngDecParseHeader", libPngDec_scePngDecParseHeader)
	elf.RegisterStub("libScePngDec", "scePngDecDecode", libPngDec_scePngDecDecode)
	elf.RegisterStub("libScePngDec", "scePngDecDelete", libPngDec_scePngDecDelete)
}

func libPngDec_scePngDecQueryMemorySize(paramPtr uintptr) uintptr {
	logger.Printf("%-132s %s called.\n", emu.GlobalModuleManager.GetCallSiteText(), color.Magenta.Sprint("scePngDecQueryMemorySize"))
	return 0x20000
}

func libPngDec_scePngDecCreate(paramPtr uintptr, memoryAddress uintptr, memorySize uintptr, handlePtr uintptr) uintptr {
	logger.Printf("%-132s %s called.\n", emu.GlobalModuleManager.GetCallSiteText(), color.Magenta.Sprint("scePngDecCreate"))
	if handlePtr != 0 {
		*(*uintptr)(unsafe.Pointer(handlePtr)) = 1
	}
	return 0
}

func libPngDec_scePngDecParseHeader(paramPtr uintptr, imageInfoPtr uintptr) uintptr {
	param := (*OrbisPngDecParseParam)(unsafe.Pointer(paramPtr))
	info := (*OrbisPngDecImageInfo)(unsafe.Pointer(imageInfoPtr))

	pngData := unsafe.Slice((*byte)(unsafe.Pointer(param.PngMemAddr)), param.PngMemSize)
	config, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		return 0x80000001
	}

	// Populate the struct the game provided.
	info.ImageWidth = uint32(config.Width)
	info.ImageHeight = uint32(config.Height)

	logger.Printf("%-132s %s called.\n", emu.GlobalModuleManager.GetCallSiteText(), color.Magenta.Sprint("scePngDecParseHeader"))
	return 0
}

func libPngDec_scePngDecDecode(handle uintptr, paramPtr uintptr, imageInfoPtr uintptr) uintptr {
	param := (*PngDecDecodeParam)(unsafe.Pointer(paramPtr))
	if param.ImageMemAddr == 0 || param.ImageMemSize == 0 {
		return 0x80000002
	}
	logger.Printf("%-132s %s decoding png from 0x%X (size %d) to 0x%X (size %d).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePngDecDecode"),
		param.PngMemAddr, param.PngMemSize,
		param.ImageMemAddr, param.ImageMemSize,
	)

	pngData := unsafe.Slice((*byte)(unsafe.Pointer(param.PngMemAddr)), param.PngMemSize)
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		logger.Printf("Failed to decode PNG: %v\n", err)
		return ERR_PTR
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Copy pixels to destination memory.
	dstSlice := unsafe.Slice((*byte)(unsafe.Pointer(param.ImageMemAddr)), param.ImageMemSize)

	pitch := int(param.ImagePitch)
	if pitch == 0 {
		pitch = width * 4
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 16-bit to 8-bit.
			r8 := byte(r >> 8)
			g8 := byte(g >> 8)
			b8 := byte(b >> 8)
			a8 := byte(a >> 8)

			dstOffset := y*pitch + x*4
			if dstOffset+3 < len(dstSlice) {
				dstSlice[dstOffset+0] = r8
				dstSlice[dstOffset+1] = g8
				dstSlice[dstOffset+2] = b8
				dstSlice[dstOffset+3] = a8
			}
		}
	}

	return 0
}

func libPngDec_scePngDecDelete(handle uintptr) uintptr {
	logger.Printf("%-132s %s called.\n", emu.GlobalModuleManager.GetCallSiteText(), color.Magenta.Sprint("scePngDecDelete"))
	return 0
}
