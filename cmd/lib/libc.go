package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterLibcStubs() {
	// CXA guard functions.
	elf.RegisterStub("libc", "__cxa_guard_release", libLibc___cxa_guard_release)
}
