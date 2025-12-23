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
func libKernel_ipmimgr_call(op, handle, resultPtr, paramsPtr, paramsSize, magic, objPtr, namePtr, configPtr uintptr) uintptr {
	if (op == IMPI_CREATE_CLIENT || op == IMPI_CREATE_SERVER || op == IMPI_CONNECT || op == IMPI_DISCONNECT) && magic != IPMI_MAGIC {
		fmt.Printf("%-120s %s calling using invalid magic %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", magic),
		)
	}

	var client *IpmiClient
	if op != IMPI_CREATE_CLIENT && op != IMPI_CREATE_SERVER && op != IMPI_DESTROY {
		client = GetImpiClient(uint32(handle))
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
	case IMPI_CREATE_CLIENT:
		name := "unnamed"
		if namePtr != 0 {
			name = ReadCString(namePtr)
		}

		client = CreateImpiClient(name, objPtr)
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, client.Handle)
		}

		fmt.Printf("%-120s %s created ipmi client %s (name=%s, objPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", client.Handle),
			color.Blue.Sprintf(client.Name),
			color.Yellow.Sprintf("0x%X", client.ObjPtr),
		)
		return 0

	case IMPI_CREATE_SERVER:
		name := "unnamed"
		if namePtr != 0 {
			name = ReadCString(namePtr)
		}

		server := CreateImpiServer(name, objPtr)
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, server.Handle)
		}

		fmt.Printf("%-120s %s created ipmi server %s (name=%s, objPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", server.Handle),
			color.Blue.Sprintf(server.Name),
			color.Yellow.Sprintf("0x%X", server.ObjPtr),
		)
		return 0

	case IMPI_DESTROY:
		delete(GlobalIpmiManager.Clients, uint32(handle))
		delete(GlobalIpmiManager.Servers, uint32(handle))
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		fmt.Printf("%-120s %s destroyed ipmi object %s.\n",
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
		if _, err := GlobalFilesystem.Write(fmt.Sprintf("/%s", client.Name), make([]byte, ImpiBufferDefault)); err != nil {
			fmt.Printf("%-120s %s failed creating shared memory file (%s).\n",
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

		fmt.Printf("%-120s %s connected %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return 0

	case IMPI_DISCONNECT:
		if err := GlobalFilesystem.Delete(fmt.Sprintf("/%s", client.Name)); err != nil {
			fmt.Printf("%-120s %s failed deleting shared memory file (%s).\n",
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

		if objPtr != 0 {
			objSlice := unsafe.Slice((*byte)(unsafe.Pointer(objPtr)), 4)
			binary.LittleEndian.PutUint32(objSlice, 0)
		}

		fmt.Printf("%-120s %s disconnected %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", handle),
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
