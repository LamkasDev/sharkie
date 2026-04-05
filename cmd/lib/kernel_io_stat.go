package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x00000000000163D0
// __int64 __fastcall sceKernelStat(__int64, __int64)
func libKernel_sceKernelStat(pathPtr Cstring, statPtr uintptr) int32 {
	err := libKernel_stat(pathPtr, statPtr)
	if err != 0 {
		return int32(GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000000850
// __int64 __fastcall stat()
func libKernel_stat(pathPtr Cstring, statPtr uintptr) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
		)
		return 0
	}

	path := GetUsablePath(GoString(pathPtr))
	fd, err := GlobalFilesystem.Open(path, 0, 0)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	return libKernel_fstat(fd, statPtr)
}

// 0x0000000000016400
// __int64 __fastcall sceKernelFstat(__int64, __int64)
func libKernel_sceKernelFstat(fd FileDescriptor, statPtr uintptr) int32 {
	err := libKernel_fstat(fd, statPtr)
	if err != 0 {
		return int32(GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x00000000000009D0
// __int64 __fastcall fstat()
func libKernel_fstat(fd FileDescriptor, statPtr uintptr) int32 {
	if statPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid stat pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	fileStat, err := file.File.Stat()
	if err != nil {
		logger.Printf("%-132s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	stat := (*FileStat)(unsafe.Pointer(statPtr))
	stat.Device = 0
	stat.Inodes = 0
	stat.Mode = 0
	stat.HardLinkCount = 1
	stat.OwnerUser = 0
	stat.OwnerGroup = 0
	stat.SpecialDevice = 0
	stat.AccessTime = Timestamp{Seconds: 0, Nanoseconds: 0}
	stat.ModifyTime = Timestamp{Seconds: 0, Nanoseconds: 0}
	stat.ChangeStatusTime = Timestamp{Seconds: 0, Nanoseconds: 0}
	stat.Size = fileStat.Size()
	stat.BlockSize = FileBlockSize
	stat.Blocks = (stat.Size + 511) / 512
	stat.Flags = 0
	stat.GenerationNumber = 0
	stat.ImplementationDetails = 0
	stat.CreateTime = Timestamp{Seconds: 0, Nanoseconds: 0}

	logger.Printf("%-132s %s returned file stat for %s (size=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("fstat"),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", stat.Size),
	)
	return 0
}
