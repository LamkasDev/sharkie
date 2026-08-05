package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Getpagesize() uintptr {
	return libScePosix_getpagesize()
}

func libScePosix_getpagesize() uintptr {
	return uintptr(MemoryPageSize)
}

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
	if isFixed && addr != 0 && addr%uintptr(MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to invalid MAP_FIXED address alignment %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		logger.Printf(emu.SprintStackTrace())
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(uintptr(MemoryPageSize) - 1)
	alignedSize := (length + MemoryPageSize - 1) & ^(MemoryPageSize - 1)

	// Enforce file descriptor minimum length.
	hasFile := !isAnonymous && fd != ERR_PTRI && uint32(fd) != ERR_HANDLE
	if hasFile {
		if alignedSize < MinFileMmapSize {
			logger.Printf("%-132s %s expanding allocation size from %s to %s bytes.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap"),
				color.Yellow.Sprintf("0x%X", alignedSize),
				color.Yellow.Sprintf("0x%X", MinFileMmapSize),
			)
			alignedSize = MinFileMmapSize
		}
	}

	// If we need to write file contents into the block, we must set the PROT_WRITE flag temporarily.
	tempProt := prot
	if hasFile {
		tempProt |= PROT_WRITE
	}

	// Allocate memory and check error.
	allocatedAddr, err := AllocKernelMemory(alignedAddr, alignedSize, tempProt, flags, uintptr(MemoryPageSize))
	if allocatedAddr == 0 {
		// If we're not required to return a fixed address, try again without the hint.
		if !isFixed && alignedAddr != 0 {
			allocatedAddr, err = AllocKernelMemory(0, alignedSize, tempProt, flags, uintptr(MemoryPageSize))
		}
	}
	if allocatedAddr == 0 {
		logger.Printf("%-132s %s failed allocating memory (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s, err=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", alignedAddr),
			color.Yellow.Sprintf("0x%X", alignedSize),
			color.Blue.Sprint(MemoryProtName(prot)),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", offset),
			err.Error(),
		)
		emu.SetErrno(ENOMEM)
		return ERR_PTR
	}
	if !isFixed && alignedAddr != 0 && allocatedAddr != alignedAddr {
		logger.Printf("%-132s %s ignored allocation address hint (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap"),
			color.Yellow.Sprintf("0x%X", alignedAddr),
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
			end := int(offset) + int(alignedSize)
			if end > len(fileData) {
				end = len(fileData)
			}
			fileChunk := fileData[int(offset):end]

			memorySlice := unsafe.Slice((*byte)(unsafe.Pointer(allocatedAddr)), len(fileChunk))
			copy(memorySlice, fileChunk)
		}

		// Protect the memory block again to its requested state.
		if tempProt != prot {
			if _, err = ProtectKernelMemory(allocatedAddr, alignedSize, prot); err != nil {
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

	HookMap(allocatedAddr, alignedSize, prot)

	logger.Printf("%-132s %s allocated %s bytes at %s (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mmap"),
		color.Yellow.Sprintf("0x%X", alignedSize),
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
