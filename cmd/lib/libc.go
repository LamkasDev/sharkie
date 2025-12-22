package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterLibcStubs() {
	// Memory functions.
	elf.RegisterStub("libc", "malloc", libc_malloc)
	elf.RegisterStub("libc", "calloc", libc_calloc)
	elf.RegisterStub("libc", "free", libc_free)
	elf.RegisterStub("libc", "realloc", libc_realloc)
	elf.RegisterStub("libc", "sceLibcMspaceMalloc", libc_sceLibcMspaceMalloc)
	elf.RegisterStub("libc", "sceLibcMspaceCalloc", libc_sceLibcMspaceCalloc)
	elf.RegisterStub("libc", "sceLibcMspaceFree", libc_sceLibcMspaceFree)
	elf.RegisterStub("libc", "sceLibcMspaceRealloc", libc_sceLibcMspaceRealloc)

	// CXA guard functions.
	elf.RegisterStub("libc", "__cxa_guard_release", libLibc___cxa_guard_release)
}
