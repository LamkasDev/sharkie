package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Munmap(addr uintptr, length uint64) uintptr {
	return libScePosix_munmap(addr, length)
}

func libScePosix_munmap(addr uintptr, length uint64) uintptr {
	// Perform initial pointer checks.
	if addr == 0 || length == 0 {
		logger.Printf("%-132s %s failed due to invalid address %s or size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("munmap"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", length),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Check adress alignment, round down size to page boundary.
	if addr%uintptr(MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to unaligned address %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("munmap"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(uintptr(MemoryPageSize) - 1)
	alignedSize := (length + MemoryPageSize - 1) & ^(MemoryPageSize - 1)

	_, err := FreeKernelMemory(alignedAddr, alignedSize)
	if err != nil {
		logger.Printf("%-132s %s failed due to unmap error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("munmap"),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	HookUnmap(alignedAddr, uintptr(alignedSize))

	logger.Printf("%-132s %s unmapped %s bytes at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("munmap"),
		color.Yellow.Sprintf("0x%X", alignedAddr),
		color.Yellow.Sprintf("0x%X", alignedSize),
	)
	return 0
}
