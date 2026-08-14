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
		return 0x805A1000
	}
	if module := emu.GlobalModuleManager.GetModule(moduleName); module != nil {
		return uintptr(module.ModuleIndex)
	}

	return 0x805A1001
}

// 0x0000000000000410
// __int64 __fastcall sceSysmoduleLoadModule(unsigned int)
func libSceSysmodule_sceSysmoduleLoadModule(id SysmoduleId) uintptr {
	return libSceSysmodule_sceSysmoduleLoadModuleInternal(id)
}

// 0x00000000000005D0
// __int64 __fastcall sceSysmoduleLoadModuleInternal(__int64)
func libSceSysmodule_sceSysmoduleLoadModuleInternal(id SysmoduleId) uintptr {
	return libSceSysmodule_sceSysmoduleLoadModuleInternalWithArg(id, 0, 0, 0, nil)
}

// 0x0000000000000710
// __int64 __fastcall sceSysmoduleLoadModuleInternalWithArg(__int64, __int64, __int64, int, __int64)
func libSceSysmodule_sceSysmoduleLoadModuleInternalWithArg(id SysmoduleId, argc, argvPtr uintptr, idk uint64, resultPtr *int32) uintptr {
	moduleName, ok := SysmoduleMap[id]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown module id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModuleInternalWithArg"),
			color.Green.Sprint(id),
		)
		return 0x805A1000
	}
	if emu.GlobalModuleManager.IsModuleLoaded(moduleName) {
		logger.Printf("%-132s %s skipping already loaded module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModuleInternalWithArg"),
			color.Blue.Sprint(moduleName),
		)
		return 0
	}
	logger.Printf("%-132s %s loading %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSysmoduleLoadModuleInternalWithArg"),
		color.Blue.Sprint(moduleName),
	)

	// Load the module and its dependencies into memory and link them.
	mod, err := emu.GlobalModuleManager.LoadModule(moduleName, true)
	if err != nil {
		logger.Printf("%-132s %s failed due to load error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSysmoduleLoadModuleInternalWithArg"),
			err.Error(),
		)
		return 0x80020000
	}

	// Initialize the module and its dependencies synchronously using CallAndWait.
	ret := emu.GlobalModuleManager.RunModuleInitializers(emu.GetCurrentThread(), mod, true, false, argc, argvPtr)

	// Write back return value of init function.
	if resultPtr != nil {
		*resultPtr = int32(ret)
	}

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
		return 0x805A1000
	}
	logger.Printf("%-132s %s tried unloading %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSysmoduleUnloadModule"),
		color.Blue.Sprint(moduleName),
	)

	return 0
}
