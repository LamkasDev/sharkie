package libc

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// TODO: this needs buffering, proper mode parsing + other fun stuff.
// 0x000000000000A5F0
// __int64 __fastcall fopen(__int64, __int64)
func libSceLibcInternal_fopen(pathPtr, modePtr Cstring) uintptr {
	fd := posix.Open(pathPtr, O_RDWR, 0777)
	if fd == ERR_PTRI {
		return 0
	}

	return uintptr(fd)
}

// TODO: this
// 0x000000000000AC00
// unsigned __int64 __fastcall fread(_BYTE *, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_fread(ptr, size, n, filePtr uintptr) uintptr {
	read := posix.Read(FileDescriptor(filePtr), ptr, uint64(size*n))
	if read == ERR_PTRI {
		return 0
	}

	return uintptr(read)
}

// TODO: this
// 0x000000000000AF00
// __int64 __fastcall fseek(__int64, __int64, unsigned int)
func libSceLibcInternal_fseek(filePtr, offset, whence uintptr) uintptr {
	return uintptr(posix.Lseek(FileDescriptor(filePtr), int64(offset), int32(whence)))
}

// TODO: this
// 0x000000000000A1B0
// __int64 __fastcall fgetpos(__int64, __int64)
func libSceLibcInternal_fgetpos(filePtr, posPtr uintptr) uintptr {
	offset, err := GlobalFilesystem.GetOffsetFd(FileDescriptor(filePtr))
	if err != nil {
		logger.Printf("%-132s %s failed due to get offset error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
			color.Yellow.Sprintf("0x%X", filePtr),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else {
			emu.SetErrno(ESPIPE)
		}
		return ERR_PTR
	}
	WriteAddress(posPtr, uintptr(offset))

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned %s cursor of %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek_0"),
			color.Yellow.Sprintf("0x%X", filePtr),
			color.Yellow.Sprintf("0x%X", offset),
		)
	}
	return 0
}

// TODO: this
// 0x0000000000010720
// __int64 __fastcall setvbuf(__int16 *, __int64, int, unsigned __int64)
func libSceLibcInternal_setvbuf(filePtr, bufferPtr, mode, size uintptr) uintptr {
	return 0
}

// TODO: this
// 0x0000000000009C50
// __int64 __fastcall fclose(__int64)
func libSceLibcInternal_fclose(filePtr uintptr) uintptr {
	return uintptr(posix.Close(FileDescriptor(filePtr)))
}
