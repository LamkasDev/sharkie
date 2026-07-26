package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Mmap(addr uintptr, length uint64, prot, flags int32, fd FileDescriptor, offset uintptr) uintptr {
	return libScePosix_mmap(addr, length, prot, flags, fd, offset)
}

func libScePosix_mmap(addr uintptr, length uint64, prot, flags int32, fd FileDescriptor, offset uintptr) uintptr {
	// Perform initial pointer checks.
	if length == 0 {
		logger.Printf("%-132s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", length),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Check offset alignment.
	if offset%uintptr(MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to unaligned offset %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", offset),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Check adress alignment, round down hint to page boundary.
	isFixed := (flags & MAP_FIXED) != 0
	isAnonymous := (flags & MAP_ANON) != 0
	if addr != 0 {
		if addr%uintptr(MemoryPageSize) != 0 {
			if isFixed {
				logger.Printf("%-132s %s failed due to invalid MAP_FIXED address alignment %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("mmap"),
					color.Yellow.Sprintf("0x%X", addr),
				)
				logger.Printf(emu.SprintStackTrace())
				emu.SetErrno(EINVAL)
				return ERR_PTR
			}
			addr = addr & ^(uintptr(MemoryPageSize) - 1)
		}
	}

	// Round up allocation length, enforce file descriptor minimum length.
	allocatedLength := uint64((length + uint64(MemoryPageSize) - 1) & ^(uint64(MemoryPageSize) - 1))
	hasFile := !isAnonymous && fd != ERR_PTRI && uint32(fd) != ERR_HANDLE
	if hasFile {
		if allocatedLength < MinFileMmapSize {
			logger.Printf("%-132s %s expanding allocation size from %s to %s bytes.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap"),
				color.Yellow.Sprintf("0x%X", allocatedLength),
				color.Yellow.Sprintf("0x%X", MinFileMmapSize),
			)
			allocatedLength = MinFileMmapSize
		}
	}

	// If we need to write file contents into the block, we must set the PROT_WRITE flag temporarily.
	tempProt := prot
	if hasFile {
		tempProt |= PROT_WRITE
	}

	// Allocate memory and check error.
	allocatedAddr, err := AllocKernelMemory(addr, allocatedLength, tempProt, flags)
	if allocatedAddr == 0 {
		// If we're not required to return a fixed address, try again without the hint.
		if !isFixed && addr != 0 {
			allocatedAddr, err = AllocKernelMemory(0, allocatedLength, tempProt, flags)
		}
	}
	if allocatedAddr == 0 {
		logger.Printf("%-132s %s failed allocating memory (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s, err=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", length),
			color.Blue.Sprint(MemoryProtName(prot)),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", offset),
			err.Error(),
		)
		emu.SetErrno(ENOMEM)
		return ERR_PTR
	}
	if !isFixed && addr != 0 && allocatedAddr != addr {
		logger.Printf("%-132s %s ignored allocation address hint (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
		)
	}

	// Handle file descriptor copy if it's a file-backed mapping.
	if hasFile {
		file, ok := GlobalFilesystem.Descriptors[fd]
		if !ok {
			logger.Printf("%-132s %s failed due to unknown file descriptor %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap"),
				color.Yellow.Sprintf("0x%X", fd),
			)
			emu.SetErrno(EBADF)
			return ERR_PTR
		}

		// Copy file data into the memory block.
		fileData, err := GlobalFilesystem.ReadFull(file.Path)
		if err != nil {
			logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap"),
				color.Blue.Sprint(file.Path),
				err.Error(),
			)
			emu.SetErrno(EFAULT)
			return ERR_PTR
		}
		if int(offset) < len(fileData) {
			end := int(offset) + int(length)
			if end > len(fileData) {
				end = len(fileData)
			}
			fileChunk := fileData[int(offset):end]

			memorySlice := unsafe.Slice((*byte)(unsafe.Pointer(allocatedAddr)), len(fileChunk))
			copy(memorySlice, fileChunk)
		}

		// Protect the memory block again to its requested state.
		if tempProt != prot {
			if _, err = ProtectKernelMemory(allocatedAddr, allocatedLength, prot); err != nil {
				logger.Printf("%-132s %s failed due to memory protection error (%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("mmap"),
					err.Error(),
				)
				emu.SetErrno(EFAULT)
				return ERR_PTR
			}
		}
	}

	HookMap(allocatedAddr, allocatedLength, prot)

	logger.Printf("%-132s %s allocated %s bytes at %s (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mmap"),
		color.Yellow.Sprintf("0x%X", allocatedLength),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", length),
		color.Blue.Sprint(MemoryProtName(prot)),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", offset),
	)
	return allocatedAddr
}

func Munmap(addr uintptr, length uint64) uintptr {
	return libScePosix_munmap(addr, length)
}

func libScePosix_munmap(addr uintptr, length uint64) uintptr {
	if addr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("munmap"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	HookUnmap(addr, uintptr(length))

	_, err := FreeKernelMemory(addr, length)
	if err != nil {
		logger.Printf("%-132s %s failed to unmap %s (length=%s, err=%s)\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("munmap"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", length),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s unmapped %s bytes at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("munmap"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", addr),
	)
	return 0
}

func Mname(addr uintptr, length uint64, namePtr Cstring) uintptr {
	return libScePosix_mname(addr, length, namePtr)
}

func libScePosix_mname(addr uintptr, length uint64, namePtr Cstring) uintptr {
	// Perform initial pointer checks.
	if addr == 0 {
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	name := "unnamed"
	if namePtr != nil {
		name = GoString(namePtr)
	}

	// TODO: actually name the regions.
	logger.Printf("%-132s %s marked %s bytes at %s as %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mname"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", addr),
		color.Blue.Sprintf(name),
	)
	return 0
}
