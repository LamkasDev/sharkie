package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000002010
// __int64 __fastcall ipmimgr_call()
func libKernel_ipmimgr_call(op uint64, handle uint32, resultPtr uintptr, paramsPtr uintptr, paramsSize uint64, magic uint64) uintptr {
	if (op == IMPI_CREATE || op == IMPI_CONNECT) && magic != IPMI_MAGIC {
		fmt.Printf("%-120s %s calling using invalid magic %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", magic),
		)
	}

	var client *IpmiClient
	if op != IMPI_CREATE {
		client = GetImpiClient(handle)
		if client == nil {
			fmt.Printf("%-120s %s failed due to unknown handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return ENOENT
		}
	}

	switch op {
	case IMPI_CREATE:
		if paramsSize < 24 || paramsPtr == 0 {
			fmt.Printf("%-120s %s failed due to invalid params pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}

		paramsSlice := unsafe.Slice((*byte)(unsafe.Pointer(paramsPtr)), 24)
		clientImplPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice))
		namePtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[8:]))
		// configPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[16:]))

		name := "unnamed"
		if namePtr != 0 {
			name = ReadCString(namePtr)
		}

		client = CreateImpiClient(name, clientImplPtr)
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, client.Handle)
		}

		fmt.Printf("%-120s %s created ipmi client %s (name=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", client.Handle),
			color.Blue.Sprintf(client.Name),
		)
		return 0

	case IMPI_DESTROY:
		fmt.Printf("%-120s %s destroyed ipmi client %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return 0

	case IMPI_INVOKE_SYNC:
		if paramsPtr == 0 {
			fmt.Printf("%-120s %s failed due to invalid params pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}

		paramsSlice := unsafe.Slice((*byte)(unsafe.Pointer(paramsPtr)), 48)
		methodId := uintptr(binary.LittleEndian.Uint32(paramsSlice))
		// inputPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x10:]))
		// inputSizePtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x18:]))
		// outputPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x20:]))
		// outputSizePtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x28:]))

		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		fmt.Printf("%-120s %s invoked unknown method %s (handle=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", methodId),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return 0

	case IMPI_CONNECT:
		shmFile, err := GlobalShmFilesystem.CreateFile(fmt.Sprintf("/%s", client.Name))
		if err != nil {
			fmt.Printf("%-120s %s failed creating shared memory shmFile: %+v\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				err.Error(),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}
		if _, err = shmFile.Write(make([]byte, ImpiBufferDefault)); err != nil {
			fmt.Printf("%-120s %s failed writing into shared memory shmFile: %+v\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				err.Error(),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}

		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		fmt.Printf("%-120s %s connected %s (shmDescriptor=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", handle),
			color.Yellow.Sprintf("0x%X", shmFile.Descriptor),
		)
		return 0
	}

	fmt.Printf("%-120s %s failed due to unknown operation %s (handle=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ipmimgr_call"),
		color.Yellow.Sprintf("0x%X", op),
		color.Yellow.Sprintf("0x%X", handle),
	)
	return SCE_KERNEL_ERROR_ENOTSUP
}
