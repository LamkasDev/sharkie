package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000125A0
// __int64 __fastcall mmap(__int64, __int64, __int64, __int64, __int64, __int64)
func libKernel_mmap(addr, length, prot, flags, fd, offset uintptr) uintptr {
	return libKernel_mmap_0(addr, length, prot, flags, fd, offset)
}

// 0x0000000000002990
// __int64 __fastcall mmap_0()
func libKernel_mmap_0(addr, length, prot, flags, fd, offset uintptr) uintptr {
	// Perform initial pointer checks.
	if length == 0 {
		fmt.Printf("%-120s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			color.Yellow.Sprintf("0x%X", length),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	// If we need to write into the block, we need to set the flag it for a bit.
	tempProt := prot
	if fd != ERR_PTR && uint32(fd) != ERR_HANDLE {
		tempProt |= PROT_WRITE
	}

	// Libraries sometimes rely on the allocated file block being at least this size.
	allocatedLength := length
	if fd != ERR_PTR && uint32(fd) != ERR_HANDLE {
		if allocatedLength < MinFileMmapSize {
			fmt.Printf("%-120s %s expanding allocation size from %s to %s bytes.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap_0"),
				color.Yellow.Sprintf("0x%X", allocatedLength),
				color.Yellow.Sprintf("0x%X", MinFileMmapSize),
			)
			allocatedLength = MinFileMmapSize
		}
	}

	// Allocate memory and check error.
	allocatedAddr, err := AllocKernelMemory(addr, allocatedLength, tempProt, flags)
	if allocatedAddr == 0 {
		// If we're not required to return a fixed address, let's try again and let Windows choose.
		if (flags&MAP_FIXED) == 0 && addr != 0 {
			allocatedAddr, err = AllocKernelMemory(0, allocatedLength, tempProt, flags)
		}
	}
	if allocatedAddr == 0 {
		fmt.Printf("%-120s %s failed allocating memory (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s, err=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", length),
			color.Yellow.Sprintf("0x%X", prot),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", offset),
			err.Error(),
		)
		SetErrno(ENOMEM)
		return ERR_PTR
	}
	if addr != 0 && allocatedAddr != addr {
		fmt.Printf("%-120s %s ignored allocation address (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
		)
	}

	// Handle file descriptor copy.
	if fd != ERR_PTR && uint32(fd) != ERR_HANDLE {
		file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
		if !ok {
			fmt.Printf("%-120s %s failed due to unknown file %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("mmap_0"),
				color.Yellow.Sprintf("0x%X", fd),
			)
			SetErrno(ENOENT)
			return ERR_PTR
		}

		// Copy file data into the memory block.
		fileData, err := GlobalFilesystem.ReadFull(file.Path)
		if err != nil {
			fmt.Printf("%-120s %s failed due to read error on %s (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ftruncate_0"),
				color.Blue.Sprint(file.Path),
				err.Error(),
			)
			SetErrno(EFAULT)
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

		// Protect the memory block again.
		if tempProt != prot {
			if _, err = ProtectKernelMemory(allocatedAddr, allocatedLength, prot); err != nil {
				fmt.Printf("%-120s %s failed due to memory protection error (%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("mmap_0"),
					err.Error(),
				)
				SetErrno(EFAULT)
				return ERR_PTR
			}
		}
	}

	fmt.Printf("%-120s %s allocated %s bytes at %s (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mmap_0"),
		color.Yellow.Sprintf("0x%X", allocatedLength),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", prot),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", offset),
	)
	return allocatedAddr
}

// 0x0000000000016580
// __int64 __fastcall sceKernelMmap(__int64, __int64, __int64, __int64, __int64, __int64, __int64 *)
func libKernel_sceKernelMmap(addr, length, prot, flags, fd, offset, retAddrPtr uintptr) uintptr {
	allocatedAddr := libKernel_mmap(addr, length, prot, flags, fd, offset)
	if allocatedAddr == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	if retAddrPtr != 0 {
		retAddrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(retAddrPtr)), 8)
		binary.LittleEndian.PutUint64(retAddrPtrSlice, uint64(allocatedAddr))
	}

	return 0
}

// 0x00000000000149E0
// __int64 sceKernelMunmap()
func libKernel_sceKernelMunmap(addr, length uintptr) uintptr {
	err := libKernel_munmap(addr, length)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// 0x00000000000009F0
// __int64 __fastcall munmap()
func libKernel_munmap(addr, length uintptr) uintptr {
	if addr == 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMunmap"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	_, err := FreeKernelMemory(addr, length)
	if err != nil {
		fmt.Printf("%-120s %s failed to unmap %s (length=%s, err=%s)\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMunmap"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", length),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	fmt.Printf("%-120s %s unmapped %s bytes at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMunmap"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", addr),
	)
	return 0
}
