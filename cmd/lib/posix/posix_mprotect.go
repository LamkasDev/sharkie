package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Mprotect(addr uintptr, size uint64, prot int32) uintptr {
	return libScePosix_mprotect(addr, size, prot)
}

func libScePosix_mprotect(addr uintptr, size uint64, prot int32) uintptr {
	// Perform initial pointer checks.
	if addr == 0 || size == 0 {
		logger.Printf("%-132s %s failed due to invalid address %s or size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mprotect"),
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
			color.Magenta.Sprint("mprotect"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(uintptr(MemoryPageSize) - 1)
	alignedSize := (size + MemoryPageSize - 1) & ^(MemoryPageSize - 1)

	_, err := ProtectKernelMemory(alignedAddr, alignedSize, prot)
	if err != nil {
		logger.Printf("%-132s %s failed due to memory protection error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mprotect"),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	HookProtect(alignedAddr, uintptr(alignedSize), prot)

	logger.Printf("%-132s %s changed protection of %s bytes at %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mprotect"),
		color.Yellow.Sprintf("0x%X", alignedSize),
		color.Yellow.Sprintf("0x%X", alignedAddr),
		color.Blue.Sprint(MemoryProtName(prot)),
	)
	return 0
}
