package libc

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000A5F0
// __int64 __fastcall fopen(__int64, __int64)
func libSceLibcInternal_fopen(pathPtr, modePtr Cstring) uintptr {
	if pathPtr == nil || modePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fopen"),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	flags, filePerms, libcMode, err := ParseFileMode(GoString(modePtr))
	if err != nil {
		logger.Printf("%-132s %s failed due to invalid mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fopen"),
			color.Red.Sprint(GoString(modePtr)),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	fd := posix.Open(pathPtr, flags, filePerms)
	if fd == ERR_PTRI {
		return 0
	}
	fileAddress := GlobalGoAllocator.MallocAligned(LibcFileSize, 8)
	file := NewLibcFile(fileAddress)
	file.Mode = libcMode
	file.Handle = FileDescriptor(fd)
	file.Index = uint8(len(LibcFiles))
	LibcFiles = append(LibcFiles, file)

	return fileAddress
}

// 0x0000000000001120
// __int64 __fastcall fopen_s(__int64 *, __int64, _BYTE *, __int64)
func libSceLibcInternal_fopen_s(fileAddressPtr *uintptr, pathPtr, modePtr Cstring) uintptr {
	if fileAddressPtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fopen_s"),
		)
		return EINVAL
	}
	*fileAddressPtr = 0
	if pathPtr == nil || modePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fopen_s"),
		)
		return EINVAL
	}
	flags, filePerms, libcMode, err := ParseFileMode(GoString(modePtr))
	if err != nil {
		logger.Printf("%-132s %s failed due to invalid mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fopen_s"),
			color.Red.Sprint(GoString(modePtr)),
		)
		return EINVAL
	}
	fd := posix.Open(pathPtr, flags, filePerms)
	if fd == ERR_PTRI {
		errNo := emu.GetErrno()
		emu.SetErrno(0)
		if errNo == 0 {
			errNo = ENOENT
		}
		return errNo
	}
	fileAddress := GlobalGoAllocator.MallocAligned(LibcFileSize, 8)
	file := NewLibcFile(fileAddress)
	file.Mode = libcMode
	file.Handle = FileDescriptor(fd)
	file.Index = uint8(len(LibcFiles))
	LibcFiles = append(LibcFiles, file)
	*fileAddressPtr = fileAddress

	return 0
}

// 0x0000000000035B40
// _WORD *__fastcall fdopen(unsigned int, __int64)
func libSceLibcInternal_fdopen(fd FileDescriptor, modePtr Cstring) uintptr {
	if modePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fdopen"),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	_, _, libcMode, err := ParseFileMode(GoString(modePtr))
	if err != nil {
		logger.Printf("%-132s %s failed due to invalid mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fdopen"),
			color.Red.Sprint(GoString(modePtr)),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	errFl := posix.Fcntl(fd, F_GETFL, 0)
	if errFl == ERR_PTRI {
		return 0
	}
	fileAddress := GlobalGoAllocator.MallocAligned(LibcFileSize, 8)
	file := NewLibcFile(fileAddress)
	file.Mode = libcMode
	file.Handle = fd
	file.Index = uint8(len(LibcFiles))
	LibcFiles = append(LibcFiles, file)

	return fileAddress
}

// 0x000000000000AD50
// __int64 __fastcall freopen(__int64, _BYTE *, _WORD *)
func libSceLibcInternal_freopen(pathPtr, modePtr Cstring, file *LibcFile) uintptr {
	if modePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("freopen"),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	flags, filePerms, libcMode, err := ParseFileMode(GoString(modePtr))
	if err != nil {
		logger.Printf("%-132s %s failed due to invalid mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("freopen"),
			color.Red.Sprint(GoString(modePtr)),
		)
		emu.SetErrno(EINVAL)
		return 0
	}
	fd := posix.Open(pathPtr, flags, filePerms)
	if fd == ERR_PTRI {
		return 0
	}
	if file != nil && file.Handle >= 0 {
		_ = posix.Close(file.Handle)
		file.Handle = -1
	}

	// Create or reset file.
	if file == nil {
		fileAddress := GlobalGoAllocator.MallocAligned(LibcFileSize, 8)
		file = NewLibcFile(fileAddress)
		file.Index = uint8(len(LibcFiles))
		LibcFiles = append(LibcFiles, file)
	} else {
		NewLibcFile(uintptr(unsafe.Pointer(file)))
	}
	file.Mode = libcMode
	file.Handle = FileDescriptor(fd)
	file.CharBuffer = 0

	return uintptr(unsafe.Pointer(file))
}

// 0x0000000000001350
// __int64 __fastcall freopen_s(__int64 *, __int64, _BYTE *, _WORD *)
func libSceLibcInternal_freopen_s(fileAddressPtr *uintptr, pathPtr, modePtr Cstring, file *LibcFile) uintptr {
	if fileAddressPtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("freopen_s"),
		)
		return EINVAL
	}
	*fileAddressPtr = 0
	if modePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("freopen_s"),
		)
		return EINVAL
	}
	fileAddress := libSceLibcInternal_freopen(pathPtr, modePtr, file)
	if fileAddress == 0 {
		errNo := emu.GetErrno()
		if errNo == 0 {
			errNo = ENOENT
		}
		return errNo
	}
	*fileAddressPtr = fileAddress

	return 0
}

// 0x000000000000AC00
// unsigned __int64 __fastcall fread(_BYTE *, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_fread(ptr, size, n uintptr, file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fread"),
		)
		emu.SetErrno(EBADF)
		return 0
	}
	if size == 0 || n == 0 {
		return 0
	}
	if ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fread"),
		)
		emu.SetErrno(EFAULT)
		file.Mode |= M_ERR
		return 0
	}
	// Drain backing buffer, then read from filesystem.
	var totalRead uintptr
	total := size * n
	dest := ptr
	if (file.Mode & M_BACK) != 0 {
		*(*byte)(unsafe.Pointer(dest)) = file.CharBuffer
		file.Mode &= ^M_BACK
		dest++
		totalRead++
	}
	if totalRead < total {
		toRead := uint64(total - totalRead)
		read := posix.Read(file.Handle, dest, toRead)
		if read == ERR_PTRI {
			file.Mode |= M_ERR
		} else {
			totalRead += uintptr(read)
			if uint64(read) < toRead {
				file.Mode |= M_EOF
			}
		}
	}

	return totalRead / size
}

// 0x000000000000A030
// __int64 __fastcall fgetc(__int64)
func libSceLibcInternal_fgetc(file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	// Use backing buffer or read from filesystem.
	if (file.Mode & M_BACK) != 0 {
		file.Mode &= ^M_BACK
		return uintptr(file.CharBuffer)
	}
	var c byte
	read := posix.Read(file.Handle, uintptr(unsafe.Pointer(&c)), 1)
	if read == ERR_PTRI {
		file.Mode |= M_ERR
		return EOF
	}
	if read == 0 {
		file.Mode |= M_EOF
		return EOF
	}

	return uintptr(c)
}

// 0x00000000000126E0
// __int64 __fastcall ungetc(unsigned int, __int64)
func libSceLibcInternal_ungetc(c uintptr, file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ungetc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	if c == EOF || uint32(c) == 0xFFFFFFFF || (file.Mode&M_OPENR) == 0 || (file.Mode&M_BACK) != 0 {
		return EOF
	}
	file.CharBuffer = byte(c)
	file.Mode = (file.Mode & ^M_EOF) | M_BACK

	return uintptr(byte(c))
}

// 0x000000000000B1F0
// unsigned __int64 __fastcall fwrite(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_fwrite(ptr, size, n uintptr, file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fwrite"),
		)
		emu.SetErrno(EBADF)
		return 0
	}
	if size == 0 || n == 0 {
		return 0
	}
	if ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fwrite"),
		)
		emu.SetErrno(EFAULT)
		file.Mode |= M_ERR
		return 0
	}
	wrote := posix.Write(file.Handle, ptr, uint64(size*n))
	if wrote == ERR_PTRI {
		file.Mode |= M_ERR
		return 0
	}

	return uintptr(wrote) / size
}

// 0x000000000000A770
// __int64 __fastcall fputc(unsigned __int8, __int64)
func libSceLibcInternal_fputc(c uintptr, file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fputc"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	wrote := posix.Write(file.Handle, uintptr(unsafe.Pointer(&c)), 1)
	if wrote == ERR_PTRI {
		file.Mode |= M_ERR
		return EOF
	}

	return c
}

// 0x000000000000A910
// __int64 __fastcall fputs(_BYTE *, __int64)
func libSceLibcInternal_fputs(stringPtr Cstring, file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fputs"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	if stringPtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fputs"),
		)
		emu.SetErrno(EINVAL)
		file.Mode |= M_ERR
		return EOF
	}
	wrote := posix.Write(file.Handle, uintptr(unsafe.Pointer(stringPtr)), uint64(len(GoString(stringPtr))))
	if wrote == ERR_PTRI {
		file.Mode |= M_ERR
		return EOF
	}

	return 0
}

// 0x000000000000C630
// __int64 __fastcall putchar(unsigned __int8)
func libSceLibcInternal_putchar(c uintptr) uintptr {
	wrote := posix.Write(1, uintptr(unsafe.Pointer(&c)), 1)
	if wrote == ERR_PTRI {
		if len(LibcFiles) > 1 && LibcFiles[1] != nil {
			LibcFiles[1].Mode |= M_ERR
		}
		return EOF
	}

	return c
}

// 0x000000000000C650
// __int64 __fastcall puts(__int64)
func libSceLibcInternal_puts(stringPtr Cstring) uintptr {
	if stringPtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("puts"),
		)
		emu.SetErrno(EINVAL)
		if len(LibcFiles) > 1 && LibcFiles[1] != nil {
			LibcFiles[1].Mode |= M_ERR
		}
		return EOF
	}
	wrote := posix.Write(1, uintptr(unsafe.Pointer(stringPtr)), uint64(len(GoString(stringPtr))))
	if wrote == ERR_PTRI {
		if len(LibcFiles) > 1 && LibcFiles[1] != nil {
			LibcFiles[1].Mode |= M_ERR
		}
		return EOF
	}
	var newline byte = '\n'
	if posix.Write(1, uintptr(unsafe.Pointer(&newline)), 1) == ERR_PTRI {
		if len(LibcFiles) > 1 && LibcFiles[1] != nil {
			LibcFiles[1].Mode |= M_ERR
		}
		return EOF
	}

	return 0
}

// 0x0000000000009DE0
// __int64 __fastcall fflush(__int16 *)
func libSceLibcInternal_fflush(file *LibcFile) uintptr {
	if file == nil {
		for _, f := range LibcFiles {
			if f != nil && (f.Mode&M_ACT) != 0 && (f.Mode&(M_OPENW|M_OPENA)) != 0 {
				// Unbuffered, nothing to flush.
			}
		}
		return 0
	}
	/* if (file.Mode & M_ACT) == 0 {
		logger.Printf("%-132s %s failed due to invalid file mode.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fflush"),
		)
		emu.SetErrno(EBADF)
		return EOF
	} */

	return 0
}

// 0x000000000000AF00
// __int64 __fastcall fseek(__int64, __int64, unsigned int)
func libSceLibcInternal_fseek(file *LibcFile, offset, whence uintptr) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fseek"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	if whence > 2 {
		logger.Printf("%-132s %s failed due to invalid whence.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fseek"),
		)
		emu.SetErrno(EINVAL)
		return EOF
	}
	seekOffset := int64(offset)
	if whence == 1 && (file.Mode&M_BACK) != 0 {
		seekOffset--
	}
	newOffset := posix.Lseek(file.Handle, seekOffset, int32(whence))
	if newOffset == ERR_PTRI {
		return EOF
	}
	file.Mode &= ^(M_EOF | M_BACK)

	return 0
}

// 0x000000000000AFE0
// __int64 __fastcall ftell(__int64)
func libSceLibcInternal_ftell(file *LibcFile) uintptr {
	var pos uintptr
	err := libSceLibcInternal_fgetpos(file, uintptr(unsafe.Pointer(&pos)))
	if err != 0 {
		return EOF
	}

	return pos
}

// 0x000000000000A1B0
// __int64 __fastcall fgetpos(__int64, __int64)
func libSceLibcInternal_fgetpos(file *LibcFile, posPtr uintptr) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	if posPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	offset, err := GlobalFilesystem.GetOffsetFd(file.Handle)
	if err != nil {
		logger.Printf("%-132s %s failed due to get offset error on %s (handle=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
			color.Yellow.Sprintf("0x%X", file.Handle),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else {
			emu.SetErrno(ESPIPE)
		}
		return ERR_PTR
	}
	if (file.Mode&M_BACK) != 0 && offset > 0 {
		offset--
	}
	WriteAddress(posPtr, uintptr(offset))

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned %s (handle=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fgetpos"),
			color.Yellow.Sprintf("0x%X", offset),
			color.Yellow.Sprintf("0x%X", file.Handle),
		)
	}
	return 0
}

// 0x0000000000010720
// __int64 __fastcall setvbuf(__int16 *, __int64, int, unsigned __int64)
func libSceLibcInternal_setvbuf(file *LibcFile, bufferPtr, mode, size uintptr) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("setvbuf"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s set buffer to %s (handle=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("setvbuf"),
			color.Yellow.Sprintf("0x%X", bufferPtr),
			color.Yellow.Sprintf("0x%X", file.Handle),
		)
	}
	return 0
}

// 0x0000000000009C50
// __int64 __fastcall fclose(__int64)
func libSceLibcInternal_fclose(file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fclose"),
		)
		emu.SetErrno(EBADF)
		return EOF
	}
	res := posix.Close(file.Handle)
	file.Mode = 0
	file.Handle = -1
	if res == ERR_PTRI {
		return EOF
	}

	return 0
}

// 0x0000000000009D10
// __int64 __fastcall feof(_WORD *)
func libSceLibcInternal_feof(file *LibcFile) uintptr {
	if file == nil /* || (file.Mode&M_ACT) == 0 */ {
		logger.Printf("%-132s %s failed due to invalid file pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("feof"),
		)
		emu.SetErrno(EBADF)
		return 0
	}
	var eof uintptr
	if (file.Mode & M_EOF) != 0 {
		eof = 1
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned %s (handle=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("feof"),
			color.Green.Sprint(eof),
			color.Yellow.Sprintf("0x%X", file.Handle),
		)
	}
	return eof
}

func libSceLibcInternal__Lockfilelock(file *LibcFile) uintptr {
	return 0
}

func libSceLibcInternal__Unlockfilelock(file *LibcFile) uintptr {
	return 0
}

func libSceLibcInternal__Locksyslock() uintptr {
	return 0
}

func libSceLibcInternal__Unlocksyslock() uintptr {
	return 0
}
