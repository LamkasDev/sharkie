package libc

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSceLibcInternalStubs() {
	elf.RegisterStub("libSceLibcInternal", "abort", Abort)
	elf.RegisterStub("libSceLibcInternal", "exit", Exit)

	// Memory functions.
	elf.RegisterStub("libSceLibcInternal", "_malloc_init", libc__malloc_init)
	elf.RegisterStub("libSceLibcInternal", "malloc", libSceLibcInternal_malloc)
	// elf.RegisterStub("libSceLibcInternal", "memcpy", libSceLibcInternal_memcpy)
	// elf.RegisterStub("libSceLibcInternal", "memset", libSceLibcInternal_memset)
	elf.RegisterStub("libSceLibcInternal", "calloc", libSceLibcInternal_calloc)
	elf.RegisterStub("libSceLibcInternal", "free", libSceLibcInternal_free)
	elf.RegisterStub("libSceLibcInternal", "realloc", libSceLibcInternal_realloc)
	elf.RegisterStub("libSceLibcInternal", "memalign", libSceLibcInternal_memalign)
	elf.RegisterStub("libSceLibcInternal", "aligned_alloc", libSceLibcInternal_stub)
	elf.RegisterStub("libSceLibcInternal", "reallocalign", libSceLibcInternal_stub)
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
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceMallocUsableSize", libSceLibcInternal_sceLibcMspaceMallocUsableSize)
	elf.RegisterStub("libSceLibcInternal", "sceLibcMspaceAlignedAlloc", libSceLibcInternal_stub)

	// IO functions.
	elf.RegisterStub("libSceLibcInternal", "fopen", libSceLibcInternal_fopen)
	elf.RegisterStub("libSceLibcInternal", "fopen_s", libSceLibcInternal_fopen_s)
	elf.RegisterStub("libSceLibcInternal", "fdopen", libSceLibcInternal_fdopen)
	elf.RegisterStub("libSceLibcInternal", "freopen", libSceLibcInternal_freopen)
	elf.RegisterStub("libSceLibcInternal", "freopen_s", libSceLibcInternal_freopen_s)
	elf.RegisterStub("libSceLibcInternal", "fread", libSceLibcInternal_fread)
	elf.RegisterStub("libSceLibcInternal", "fgetc", libSceLibcInternal_fgetc)
	elf.RegisterStub("libSceLibcInternal", "ungetc", libSceLibcInternal_ungetc)
	elf.RegisterStub("libSceLibcInternal", "fwrite", libSceLibcInternal_fwrite)
	elf.RegisterStub("libSceLibcInternal", "fputc", libSceLibcInternal_fputc)
	elf.RegisterStub("libSceLibcInternal", "fputs", libSceLibcInternal_fputs)
	elf.RegisterStub("libSceLibcInternal", "putc", libSceLibcInternal_fputc)
	elf.RegisterStub("libSceLibcInternal", "putchar", libSceLibcInternal_putchar)
	elf.RegisterStub("libSceLibcInternal", "puts", libSceLibcInternal_puts)
	elf.RegisterStub("libSceLibcInternal", "fflush", libSceLibcInternal_fflush)
	elf.RegisterStub("libSceLibcInternal", "fseek", libSceLibcInternal_fseek)
	elf.RegisterStub("libSceLibcInternal", "ftell", libSceLibcInternal_ftell)
	elf.RegisterStub("libSceLibcInternal", "fgetpos", libSceLibcInternal_fgetpos)
	elf.RegisterStub("libSceLibcInternal", "setvbuf", libSceLibcInternal_setvbuf)
	elf.RegisterStub("libSceLibcInternal", "fclose", libSceLibcInternal_fclose)
	elf.RegisterStub("libSceLibcInternal", "feof", libSceLibcInternal_feof)
	elf.RegisterStub("libSceLibcInternal", "_Lockfilelock", libSceLibcInternal__Lockfilelock)
	elf.RegisterStub("libSceLibcInternal", "_Unlockfilelock", libSceLibcInternal__Unlockfilelock)
	elf.RegisterStub("libSceLibcInternal", "_Locksyslock", libSceLibcInternal__Locksyslock)
	elf.RegisterStub("libSceLibcInternal", "_Unlocksyslock", libSceLibcInternal__Unlocksyslock)

	// Standard files (should remove once we have proper support).
	RegisterFileStubs("libSceLibcInternal")
	RegisterFileStubs("libc")
}

func RegisterFileStubs(libraryName string) {
	stdin := elf.RegisterVariableStub(libraryName, "_Stdin", 8)
	stdin_0 := elf.RegisterVariableStub(libraryName, "_Stdin_0", 8)
	stdin_1 := elf.RegisterVariableStub(libraryName, "_Stdin_1", 8)
	stdin_2 := elf.RegisterVariableStub(libraryName, "_Stdin_2", 8)
	stdin_3 := elf.RegisterVariableStub(libraryName, "_Stdin_2", 8)
	WriteAddress(stdin.Address, 0)
	WriteAddress(stdin_0.Address, 0)
	WriteAddress(stdin_1.Address, 0)
	WriteAddress(stdin_2.Address, 0)
	WriteAddress(stdin_3.Address, 0)

	stdout := elf.RegisterVariableStub(libraryName, "_Stdout", 8)
	stdout_0 := elf.RegisterVariableStub(libraryName, "_Stdout_0", 8)
	stdout_1 := elf.RegisterVariableStub(libraryName, "_Stdout_1", 8)
	stdout_2 := elf.RegisterVariableStub(libraryName, "_Stdout_2", 8)
	WriteAddress(stdout.Address, 1)
	WriteAddress(stdout_0.Address, 1)
	WriteAddress(stdout_1.Address, 1)
	WriteAddress(stdout_2.Address, 1)

	stderr := elf.RegisterVariableStub(libraryName, "_Stderr", 8)
	stderr_0 := elf.RegisterVariableStub(libraryName, "_Stderr_0", 8)
	stderr_1 := elf.RegisterVariableStub(libraryName, "_Stderr_1", 8)
	stderr_2 := elf.RegisterVariableStub(libraryName, "_Stderr_2", 8)
	WriteAddress(stderr.Address, 2)
	WriteAddress(stderr_0.Address, 2)
	WriteAddress(stderr_1.Address, 2)
	WriteAddress(stderr_2.Address, 2)
}

func libSceLibcInternal_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}

func Exit(code uintptr) uintptr {
	logger.Printf(
		"%-132s exited with %s :c\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Yellow.Sprintf("0x%X", code),
	)
	logger.CleanupAndExit()

	return 0
}

func Abort() uintptr {
	logger.Printf(
		"%-132s aborted :c\n",
		emu.GlobalModuleManager.GetCallSiteText(),
	)
	logger.CleanupAndExit()

	return 0
}
