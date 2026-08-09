package video_out

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterVideoOutStubs() {
	// Setup functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutOpen", libSceVideoOut_sceVideoOutOpen)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutAdjustColor_", libSceVideoOut_stub)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutColorSettingsSetGamma_", libSceVideoOut_stub)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutConfigureOutputMode_", libSceVideoOut_stub)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutModeSetAny_", libSceVideoOut_stub)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetResolutionStatus", libSceVideoOut_sceVideoOutGetResolutionStatus)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetDeviceCapabilityInfo_", libSceVideoOut_sceVideoOutGetDeviceCapabilityInfo_)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSetWindowModeMargins", libSceVideoOut_stub)

	// Flip functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutAddFlipEvent", libSceVideoOut_sceVideoOutAddFlipEvent)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSetFlipRate", libSceVideoOut_sceVideoOutSetFlipRate)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSubmitFlip", libSceVideoOut_sceVideoOutSubmitFlip)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSubmitEopFlip", libSceVideoOut_sceVideoOutSubmitEopFlip)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetFlipStatus", libSceVideoOut_sceVideoOutGetFlipStatus)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutIsFlipPending", libSceVideoOut_sceVideoOutIsFlipPending)

	// V-blank functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutAddVblankEvent", libSceVideoOut_sceVideoOutAddVblankEvent)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetVblankStatus", libSceVideoOut_sceVideoOutGetVblankStatus)

	// Event functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetEventId", libSceVideoOut_sceVideoOutGetEventId)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetEventData", libSceVideoOut_sceVideoOutGetEventData)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetEventCount", libSceVideoOut_sceVideoOutGetEventCount)

	// Buffer functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutRegisterBuffers", libSceVideoOut_sceVideoOutRegisterBuffers)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutRegisterBufferAttribute", libSceVideoOut_sceVideoOutRegisterBufferAttribute)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSetBufferAttribute", libSceVideoOut_sceVideoOutSetBufferAttribute)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetBufferLabelAddress", libSceVideoOut_sceVideoOutGetBufferLabelAddress)
}

func libSceVideoOut_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
