package libc

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func getFileDescriptor(filePtr uintptr) FileDescriptor {
	if filePtr < 0x1000 {
		return FileDescriptor(filePtr)
	}
	val := *(*uint32)(unsafe.Pointer(filePtr))
	switch val {
	case 0x401:
		return 0 // stdin
	case 0x10402:
		return 1 // stdout
	case 0x20402:
		return 2 // stderr
	}

	return FileDescriptor(*(*uint64)(unsafe.Pointer(filePtr)))
}

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

// 0x0000000000035B40
// _WORD *__fastcall fdopen(unsigned int, __int64)
func libSceLibcInternal_fdopen(fd FileDescriptor, modePtr Cstring) uintptr {
	err := posix.Fcntl(fd, F_GETFL, 0)
	if err == ERR_PTRI {
		return 0
	}

	return uintptr(fd)
}

// 0x000000000000AC00
// unsigned __int64 __fastcall fread(_BYTE *, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_fread(ptr, size, n, filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fread"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	read := posix.Read(getFileDescriptor(filePtr), ptr, uint64(size*n))
	if read == ERR_PTRI {
		return 0
	}

	return uintptr(read) / size
}

// 0x000000000000A030
// __int64 __fastcall fgetc(__int64)
func libSceLibcInternal_fgetc(filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	var c byte
	read := posix.Read(getFileDescriptor(filePtr), uintptr(unsafe.Pointer(&c)), 1)
	if read == 0 {
		return EOF
	}

	return uintptr(c)
}

// 0x00000000000126E0
// __int64 __fastcall ungetc(unsigned int, __int64)
func libSceLibcInternal_ungetc(c, filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ungetc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	err := posix.Lseek(getFileDescriptor(filePtr), -1, 1)
	if err == ERR_PTRI {
		return EOF
	}

	return c
}

// 0x000000000000B1F0
// unsigned __int64 __fastcall fwrite(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_fwrite(ptr, size, n, filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fwrite"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	wrote := posix.Write(getFileDescriptor(filePtr), ptr, uint64(size*n))
	if wrote == ERR_PTRI {
		return 0
	}

	return uintptr(wrote) / size
}

// 0x000000000000A770
// __int64 __fastcall fputc(unsigned __int8, __int64)
func libSceLibcInternal_fputc(c, filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fputc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	wrote := posix.Write(getFileDescriptor(filePtr), uintptr(unsafe.Pointer(&c)), 1)
	if wrote == ERR_PTRI {
		return EOF
	}

	return c
}

// 0x000000000000A910
// __int64 __fastcall fputs(_BYTE *, __int64)
func libSceLibcInternal_fputs(sPtr Cstring, filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fputs"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	wrote := posix.Write(getFileDescriptor(filePtr), uintptr(unsafe.Pointer(sPtr)), uint64(len(GoString(sPtr))))
	if wrote == ERR_PTRI {
		return EOF
	}

	return 0
}

// 0x000000000000C630
// __int64 __fastcall putchar(unsigned __int8)
func libSceLibcInternal_putchar(c uintptr) uintptr {
	wrote := posix.Write(1, uintptr(unsafe.Pointer(&c)), 1)
	if wrote == ERR_PTRI {
		return EOF
	}

	return c
}

// 0x000000000000C650
// __int64 __fastcall puts(__int64)
func libSceLibcInternal_puts(sPtr Cstring) uintptr {
	wrote := posix.Write(1, uintptr(unsafe.Pointer(sPtr)), uint64(len(GoString(sPtr))))
	if wrote == ERR_PTRI {
		return EOF
	}

	return 0
}

// 0x0000000000009DE0
// __int64 __fastcall fflush(__int16 *)
func libSceLibcInternal_fflush(filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fflush"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	return 0
}

// 0x000000000000AF00
// __int64 __fastcall fseek(__int64, __int64, unsigned int)
func libSceLibcInternal_fseek(filePtr, offset, whence uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fseek"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	return uintptr(posix.Lseek(getFileDescriptor(filePtr), int64(offset), int32(whence)))
}

// 0x000000000000AFE0
// __int64 __fastcall ftell(__int64)
func libSceLibcInternal_ftell(filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftell"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	var pos uintptr
	err := libSceLibcInternal_fgetpos(filePtr, uintptr(unsafe.Pointer(&pos)))
	if err == ERR_PTR {
		return ERR_PTR
	}

	return pos
}

// 0x000000000000A1B0
// __int64 __fastcall fgetpos(__int64, __int64)
func libSceLibcInternal_fgetpos(filePtr, posPtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	offset, err := GlobalFilesystem.GetOffsetFd(getFileDescriptor(filePtr))
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
		logger.Printf("%-132s %s returned %s (filePtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
			color.Yellow.Sprintf("0x%X", offset),
			color.Yellow.Sprintf("0x%X", filePtr),
		)
	}
	return 0
}

// 0x0000000000010720
// __int64 __fastcall setvbuf(__int16 *, __int64, int, unsigned __int64)
func libSceLibcInternal_setvbuf(filePtr, bufferPtr, mode, size uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("setvbuf"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s set buffer to %s (filePtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("setvbuf"),
			color.Yellow.Sprintf("0x%X", bufferPtr),
			color.Yellow.Sprintf("0x%X", filePtr),
		)
	}
	return 0
}

// 0x0000000000009C50
// __int64 __fastcall fclose(__int64)
func libSceLibcInternal_fclose(filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fclose"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	return uintptr(posix.Close(getFileDescriptor(filePtr)))
}

// 0x0000000000009D10
// __int64 __fastcall feof(_WORD *)
func libSceLibcInternal_feof(filePtr uintptr) uintptr {
	if filePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fclose"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	eof := uintptr(0)
	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned %s (filePtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
			color.Green.Sprint(eof),
			color.Yellow.Sprintf("0x%X", filePtr),
		)
	}
	return eof
}

func libSceLibcInternal__Lockfilelock(filePtr uintptr) uintptr {
	return 0
}

func libSceLibcInternal__Unlockfilelock(filePtr uintptr) uintptr {
	return 0
}
