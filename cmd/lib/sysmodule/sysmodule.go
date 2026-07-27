package sysmodule

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterSysmoduleStubs() {
	// Load functions.
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleIsLoaded", libSceSysmodule_sceSysmoduleIsLoaded)
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleLoadModule", libSceSysmodule_sceSysmoduleLoadModule)
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleUnloadModule", libSceSysmodule_sceSysmoduleUnloadModule)
}
