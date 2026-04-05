package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x00000000000165E0
// __int64 sceKernelTruncate()
func libKernel_sceKernelTruncate(pathPtr Cstring, length int64) int32 {
	err := libKernel_truncate(pathPtr, length)
	if err != 0 {
		return int32(GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x00000000000125E0
// __int64 truncate()
func libKernel_truncate(pathPtr Cstring, length int64) int32 {
	return libKernel_truncate_0(pathPtr, length)
}

// 0x00000000000029F0
// __int64 truncate_0()
func libKernel_truncate_0(pathPtr Cstring, length int64) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate_0"),
		)
		return 0
	}

	path := GetUsablePath(GoString(pathPtr))
	fd, err := GlobalFilesystem.Open(path, 0, 0)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate_0"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	file, ok := GlobalFilesystem.Descriptors[fd]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	err = file.File.Truncate(length)
	if err != nil {
		logger.Printf("%-132s %s failed due to truncate error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	logger.Printf("%-132s %s truncated %s to %s bytes.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("truncate_0"),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", length),
	)
	return 0
}

// 0x0000000000016610
// __int64 sceKernelFtruncate()
func libKernel_sceKernelFtruncate(fd FileDescriptor, length int64) int32 {
	err := libKernel_ftruncate(fd, length)
	if err != 0 {
		return int32(GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000012580
// __int64 ftruncate()
func libKernel_ftruncate(fd FileDescriptor, length int64) int32 {
	return libKernel_ftruncate_0(fd, length)
}

// 0x0000000000002950
// __int64 ftruncate_0()
func libKernel_ftruncate_0(fd FileDescriptor, length int64) int32 {
	file, ok := GlobalFilesystem.Descriptors[fd]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	err := file.File.Truncate(length)
	if err != nil {
		logger.Printf("%-132s %s failed due to truncate error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	logger.Printf("%-132s %s truncated %s to %s bytes.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ftruncate_0"),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", length),
	)
	return 0
}
