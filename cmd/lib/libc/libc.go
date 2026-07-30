package libc

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterLibcStubs() {
	elf.RegisterStub("libc", "abort", Abort)
	elf.RegisterStub("libc", "exit", Exit)

	// Memory functions.
	elf.RegisterStub("libc", "_malloc_init", libc__malloc_init)
	elf.RegisterStub("libc", "malloc", libc_malloc)
	// elf.RegisterStub("libc", "memcpy", libc_memcpy)
	// elf.RegisterStub("libc", "memset", libc_memset)
	elf.RegisterStub("libc", "calloc", libc_calloc)
	elf.RegisterStub("libc", "free", libc_free)
	elf.RegisterStub("libc", "realloc", libc_realloc)
	elf.RegisterStub("libc", "memalign", libc_memalign)
	elf.RegisterStub("libc", "aligned_alloc", libSceLibcInternal_stub)
	elf.RegisterStub("libc", "reallocalign", libSceLibcInternal_stub)
	elf.RegisterStub("libc", "sceLibcMspaceMalloc", libc_sceLibcMspaceMalloc)
	elf.RegisterStub("libc", "sceLibcMspaceCalloc", libc_sceLibcMspaceCalloc)
	elf.RegisterStub("libc", "sceLibcMspaceFree", libc_sceLibcMspaceFree)
	elf.RegisterStub("libc", "sceLibcMspaceRealloc", libc_sceLibcMspaceRealloc)
	elf.RegisterStub("libc", "sceLibcMspaceReallocalign", libc_sceLibcMspaceReallocalign)
	elf.RegisterStub("libc", "sceLibcMspaceMemalign", libc_sceLibcMspaceMemalign)
	elf.RegisterStub("libc", "sceLibcMspacePosixMemalign", libc_sceLibcMspacePosixMemalign)
	elf.RegisterStub("libc", "sceLibcMspaceCreate", libc_sceLibcMspaceCreate)
	elf.RegisterStub("libc", "sceLibcMspaceDestroy", libc_sceLibcMspaceDestroy)
	elf.RegisterStub("libc", "sceLibcMspaceIsHeapEmpty", libc_sceLibcMspaceIsHeapEmpty)
	elf.RegisterStub("libc", "sceLibcMspaceMallocStats", libc_sceLibcMspaceMallocStats)
	elf.RegisterStub("libc", "sceLibcMspaceMallocStatsFast", libc_sceLibcMspaceMallocStatsFast)
	elf.RegisterStub("libc", "sceLibcPafMspaceIsHeapEmpty", libSceLibcInternal_stub)
	elf.RegisterStub("libc", "sceLibcMspaceAlignedAlloc", libSceLibcInternal_stub)

	// IO functions.
	elf.RegisterStub("libc", "fopen", libc_fopen)
	elf.RegisterStub("libc", "fopen_s", libc_fopen_s)
	elf.RegisterStub("libc", "fdopen", libc_fdopen)
	elf.RegisterStub("libc", "fread", libc_fread)
	elf.RegisterStub("libc", "fgetc", libc_fgetc)
	elf.RegisterStub("libc", "ungetc", libc_ungetc)
	elf.RegisterStub("libc", "fwrite", libc_fwrite)
	elf.RegisterStub("libc", "fputc", libc_fputc)
	elf.RegisterStub("libc", "fputs", libc_fputs)
	elf.RegisterStub("libc", "putc", libc_fputc)
	elf.RegisterStub("libc", "putchar", libc_putchar)
	elf.RegisterStub("libc", "puts", libc_puts)
	elf.RegisterStub("libc", "fflush", libc_fflush)
	elf.RegisterStub("libc", "fseek", libc_fseek)
	elf.RegisterStub("libc", "ftell", libc_ftell)
	elf.RegisterStub("libc", "fgetpos", libc_fgetpos)
	elf.RegisterStub("libc", "setvbuf", libc_setvbuf)
	elf.RegisterStub("libc", "fclose", libc_fclose)
	elf.RegisterStub("libc", "feof", libc_feof)
	elf.RegisterStub("libc", "_Lockfilelock", libc__Lockfilelock)
	elf.RegisterStub("libc", "_Unlockfilelock", libc__Unlockfilelock)
}
