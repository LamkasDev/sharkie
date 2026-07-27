package sysmodule

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/module"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000580
// __int64 __fastcall sceSysmoduleIsLoaded(unsigned int)
func libSceSysmodule_sceSysmoduleIsLoaded(id SysmoduleId) uintptr {
	moduleName, ok := SysmoduleMap[id]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown module id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleIsLoaded"),
			color.Green.Sprint(id),
		)
		return 0x80020000
	}
	if emu.GlobalModuleManager.IsModuleLoaded(moduleName) {
		return 0
	}

	return 0x80020017
}

// 0x0000000000000410
// __int64 __fastcall sceSysmoduleLoadModule(unsigned int)
func libSceSysmodule_sceSysmoduleLoadModule(id SysmoduleId) uintptr {
	moduleName, ok := SysmoduleMap[id]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown module id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModule"),
			color.Green.Sprint(id),
		)
		return 0x80020000
	}
	if emu.GlobalModuleManager.IsModuleLoaded(moduleName) {
		logger.Printf("%-132s %s skipping already loaded module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModule"),
			color.Blue.Sprint(moduleName),
		)
		return 0
	}
	logger.Printf("%-132s %s loading %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSysmoduleLoadModule"),
		color.Blue.Sprint(moduleName),
	)

	// Load the module and its dependencies into memory and link them.
	mod, err := emu.GlobalModuleManager.LoadModule(moduleName, true)
	if err != nil {
		logger.Printf("%-132s %s failed due to load error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModule"),
			err.Error(),
		)
		return 0x80020000
	}

	// Initialize the module and its dependencies synchronously using CallAndWait.
	emu.GlobalModuleManager.RunModuleInitializers(emu.GetCurrentThread(), mod, true, false)

	return 0
}

// 0x0000000000000520
// __int64 __fastcall sceSysmoduleUnloadModule(unsigned int)
func libSceSysmodule_sceSysmoduleUnloadModule(id SysmoduleId) uintptr {
	moduleName, ok := SysmoduleMap[id]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown module id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleIsLoaded"),
			color.Green.Sprint(id),
		)
		return 0x80020000
	}
	logger.Printf("%-132s %s tried unloading %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSysmoduleUnloadModule"),
		color.Blue.Sprint(moduleName),
	)

	return 0
}
