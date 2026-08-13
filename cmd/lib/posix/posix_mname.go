package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Mname(addr uintptr, length uint64, namePtr Cstring) uintptr {
	return libScePosix_mname(addr, length, namePtr)
}

func libScePosix_mname(addr uintptr, size uint64, namePtr Cstring) uintptr {
	// Perform initial pointer checks.
	if addr == 0 || size == 0 {
		logger.Printf("%-132s %s failed due to invalid address %s or size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mname"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", size),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Check adress alignment, round down size to page boundary.
	if addr%uintptr(MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to unaligned address %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mname"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(uintptr(MemoryPageSize) - 1)
	alignedSize := (size + MemoryPageSize - 1) & ^(MemoryPageSize - 1)

	// Name region.
	name := "unnamed"
	if namePtr != nil {
		name = GoString(namePtr)
	}
	HookName(alignedAddr, alignedSize, name)

	logger.Printf("%-132s %s marked %s bytes at %s as %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mname"),
		color.Yellow.Sprintf("0x%X", alignedSize),
		color.Yellow.Sprintf("0x%X", alignedAddr),
		color.Blue.Sprintf(name),
	)
	return 0
}
