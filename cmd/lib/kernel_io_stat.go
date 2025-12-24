package lib

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000016400
// __int64 __fastcall sceKernelFstat(__int64, __int64)
func libKernel_sceKernelFstat(fd uintptr, statPtr uintptr) uintptr {
	err := libKernel_fstat(fd, statPtr)
	if err != 0 {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// 0x00000000000009D0
// __int64 __fastcall fstat()
func libKernel_fstat(fd uintptr, statPtr uintptr) uintptr {
	if statPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid stat pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		fmt.Printf("%-120s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	fileStat, err := file.File.Stat()
	if err != nil {
		fmt.Printf("%-120s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	stat := (*FileStat)(unsafe.Pointer(statPtr))
	stat.Device = 0
	stat.Inodes = 0
	stat.Mode = 0
	stat.HardLinkCount = 1
	stat.OwnerUser = 0
	stat.OwnerGroup = 0
	stat.SpecialDevice = 0
	stat.AccessTime = Timestamp{Seconds: 0, NanoSeconds: 0}
	stat.ModifyTime = Timestamp{Seconds: 0, NanoSeconds: 0}
	stat.ChangeStatusTime = Timestamp{Seconds: 0, NanoSeconds: 0}
	stat.Size = fileStat.Size()
	stat.BlockSize = FileBlockSize
	stat.Blocks = (stat.Size + 511) / 512
	stat.Flags = 0
	stat.GenerationNumber = 0
	stat.ImplementationDetails = 0
	stat.CreateTime = Timestamp{Seconds: 0, NanoSeconds: 0}

	fmt.Printf("%-120s %s returned file stat for %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("fstat"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
	)
	return 0
}
