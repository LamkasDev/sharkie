package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
)

const (
	LibcCxaGuardMutexOffset = uintptr(0x128F30)
	LibcCxaGuardCondOffset  = uintptr(0x128F38)
)

// 0x00000000000B5990
// void _cxa_guard_release(__guard *)
func libLibc___cxa_guard_release(guardPtr uintptr) uintptr {
	emu.GlobalModuleManager.ModulesLock.RLock()
	defer emu.GlobalModuleManager.ModulesLock.RUnlock()

	module := emu.GlobalModuleManager.ModulesMap["libc.sprx"]
	return cxaGuardRelease(
		module.BaseAddress+LibcCxaGuardMutexOffset,
		module.BaseAddress+LibcCxaGuardCondOffset,
		guardPtr,
	)
}
