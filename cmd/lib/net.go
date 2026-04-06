package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterNetworkStubs() {
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetInfo", libSceNetCtl_sceNetCtlGetInfo)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetResult", libSceNetCtl_sceNetCtlGetResult)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetState", libSceNetCtl_sceNetCtlGetState)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlRegisterCallback", libSceNetCtl_sceNetCtlRegisterCallback)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlCheckCallback", libSceNetCtl_sceNetCtlCheckCallback)
}
