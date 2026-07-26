package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000163D0
// __int64 __fastcall sceKernelStat(__int64, __int64)
func libKernel_sceKernelStat(pathPtr Cstring, stat *FileStat) int32 {
	err := posix.Stat(pathPtr, stat)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000016400
// __int64 __fastcall sceKernelFstat(__int64, __int64)
func libKernel_sceKernelFstat(fd FileDescriptor, stat *FileStat) int32 {
	err := libKernel_fstat(fd, stat)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x00000000000009D0
// __int64 __fastcall fstat()
func libKernel_fstat(fd FileDescriptor, stat *FileStat) int32 {
	if stat == nil {
		logger.Printf("%-132s %s failed due to invalid stat pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	fileStat, err := GlobalFilesystem.StatFd(fd)
	if err != nil {
		logger.Printf("%-132s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else {
			emu.SetErrno(EFAULT)
		}
		return ERR_PTRI
	}
	*stat = *fileStat

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned file stat for %s (size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", stat.Size),
		)
	}
	return 0
}
