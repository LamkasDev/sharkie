package libc

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
)

func RegisterSceLibcInternalStubs() {
	elf.RegisterStub("libSceLibcInternal", "abort", Abort)

	// Memory functions.
	elf.RegisterStub("libSceLibcInternal", "_malloc_init", libc__malloc_init)
	elf.RegisterStub("libSceLibcInternal", "malloc", libSceLibcInternal_malloc)
	// elf.RegisterStub("libSceLibcInternal", "memcpy", libSceLibcInternal_memcpy)
	// elf.RegisterStub("libSceLibcInternal", "memset", libSceLibcInternal_memset)
	elf.RegisterStub("libSceLibcInternal", "calloc", libSceLibcInternal_calloc)
	elf.RegisterStub("libSceLibcInternal", "free", libSceLibcInternal_free)
	elf.RegisterStub("libSceLibcInternal", "realloc", libSceLibcInternal_realloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceMalloc", libSceLibcInternal_sceLibcMspaceMalloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceCalloc", libSceLibcInternal_sceLibcMspaceCalloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceFree", libSceLibcInternal_sceLibcMspaceFree)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceRealloc", libSceLibcInternal_sceLibcMspaceRealloc)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceReallocalign", libSceLibcInternal_sceLibcMspaceReallocalign)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceMemalign", libSceLibcInternal_sceLibcMspaceMemalign)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspacePosixMemalign", libSceLibcInternal_sceLibcMspacePosixMemalign)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceCreate", libSceLibcInternal_sceLibcMspaceCreate)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceDestroy", libSceLibcInternal_sceLibcMspaceDestroy)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceIsHeapEmpty", libSceLibcInternal_sceLibcMspaceIsHeapEmpty)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceMallocStats", libSceLibcInternal_sceLibcMspaceMallocStats)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceMallocStatsFast", libSceLibcInternal_sceLibcMspaceMallocStatsFast)
	elf.RegisterStub("libSceLibcInternal", "sceLibcPafMspaceIsHeapEmpty", libSceLibcInternal_sceLibcPafMspaceIsHeapEmpty)

	// IO functions.
	elf.RegisterStub("libSceLibcInternal", "fopen", libSceLibcInternal_fopen)
	elf.RegisterStub("libSceLibcInternal", "fread", libSceLibcInternal_fread)
	elf.RegisterStub("libSceLibcInternal", "fseek", libSceLibcInternal_fseek)
	elf.RegisterStub("libSceLibcInternal", "fgetpos", libSceLibcInternal_fgetpos)
	elf.RegisterStub("libSceLibcInternal", "setvbuf", libSceLibcInternal_setvbuf)
	elf.RegisterStub("libSceLibcInternal", "fclose", libSceLibcInternal_fclose)
}

func Abort() uintptr {
	logger.Printf(
		"%-132s aborted :c\n",
		emu.GlobalModuleManager.GetCallSiteText(),
	)
	logger.CleanupAndExit()

	return 0
}
