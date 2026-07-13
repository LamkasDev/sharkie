package invitation_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterInvitationDialogStubs() {
	elf.RegisterStub("libSceInvitationDialog", "sceInvitationDialogGetResult", libSceInvitationDialog_stub)
	elf.RegisterStub("libSceInvitationDialog", "sceInvitationDialogUpdateStatus", libSceInvitationDialog_stub)
	elf.RegisterStub("libSceInvitationDialog", "sceInvitationDialogTerminate", libSceInvitationDialog_stub)
	elf.RegisterStub("libSceInvitationDialog", "sceInvitationDialogOpen", libSceInvitationDialog_stub)
	elf.RegisterStub("libSceInvitationDialog", "sceInvitationDialogInitialize", libSceInvitationDialog_stub)
}

func libSceInvitationDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
