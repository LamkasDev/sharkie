package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterSceLibcInternalStubs() {
	elf.RegisterStub("libSceLibcInternal", "abort", Abort)

	// Memory functions.
	elf.RegisterStub("libSceLibcInternal", "malloc", libSceLibcInternal_malloc)
	elf.RegisterStub("libSceLibcInternal", "calloc", libSceLibcInternal_calloc)
	elf.RegisterStub("libSceLibcInternal", "free", libSceLibcInternal_free)
	elf.RegisterStub("libSceLibcInternal", "realloc", libSceLibcInternal_realloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceCalloc", libSceLibcInternal_sceLibcMspaceCalloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceFree", libSceLibcInternal_sceLibcMspaceFree)

	// Guard functions.
	elf.RegisterStub("libSceLibcInternal", "__cxa_guard_release", libSceLibcInternal___cxa_guard_release)

	// Mutex functions.
	elf.RegisterStub("libSceLibcInternal", "_Mtxinit", libSceLibcInternal__Mtxinit)
}
