package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterLibcStubs() {
	// Memory functions.
	elf.RegisterStub("libc", "_malloc_init", libc__malloc_init)
	elf.RegisterStub("libc", "malloc", libc_malloc)
	// elf.RegisterStub("libc", "memcpy", libc_memcpy)
	// elf.RegisterStub("libc", "memset", libc_memset)
	elf.RegisterStub("libc", "calloc", libc_calloc)
	elf.RegisterStub("libc", "free", libc_free)
	elf.RegisterStub("libc", "realloc", libc_realloc)
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

	// CXA guard functions.
	elf.RegisterStub("libc", "__cxa_guard_release", libLibc___cxa_guard_release)

	// CXA exception functions.
	elf.RegisterStub("libc", "__cxa_throw", libLibc___cxa_throw)
	elf.RegisterStub("libc", "std::_Xbad_alloc", libLibc_std_Xbad_alloc)
}
