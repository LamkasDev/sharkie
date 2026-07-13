package np_commerce

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpCommerceStubs() {
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceDialogOpen", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceDialogInitialize", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceDialogTerminate", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceDialogGetResult", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceDialogUpdateStatus", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceHidePsStoreIcon", libSceNpCommerce_stub)
	elf.RegisterStub("libSceNpCommerce", "sceNpCommerceShowPsStoreIcon", libSceNpCommerce_stub)
}

func libSceNpCommerce_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
