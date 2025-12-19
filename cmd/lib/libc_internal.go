package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterSceLibcInternalStubs() {
	// Guard functions.
	elf.RegisterStub("libSceLibcInternal", "__cxa_guard_release", libSceLibcInternal___cxa_guard_release)

	// Mutex functions.
	elf.RegisterStub("libSceLibcInternal", "_Mtxinit", libSceLibcInternal__Mtxinit)
}
