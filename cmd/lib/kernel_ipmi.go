package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000002010
// __int64 __fastcall ipmimgr_call()
func libKernel_ipmimgr_call(op, handle, resultPtr, paramsPtr, paramsSize, magic, objPtr, namePtr, configPtr uintptr) uintptr {
	if (op == IMPI_CREATE_CLIENT || op == IMPI_CREATE_SERVER) && magic != IPMI_MAGIC {
		logger.Printf("%-132s %s calling using invalid magic %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", magic),
		)
	}
	client := GetImpiClient(uint32(handle))
	server := GetImpiServer(uint32(handle))

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

		logger.Printf("%-132s %s created ipmi client %s (name=%s, objPtr=%s).\n",
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

		server = CreateImpiServer(name, objPtr)
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, server.Handle)
		}

		logger.Printf("%-132s %s created ipmi server %s (name=%s, objPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", server.Handle),
			color.Blue.Sprintf(server.Name),
			color.Yellow.Sprintf("0x%X", server.ObjPtr),
		)
		return 0

	case IMPI_DESTROY_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to invalid client handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}

		delete(GlobalIpmiManager.Clients, uint32(handle))
		delete(GlobalIpmiManager.Servers, uint32(handle))
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		logger.Printf("%-132s %s destroyed ipmi client %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_INVOKE_SYNC_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to invalid client handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}
		if paramsPtr == 0 {
			logger.Printf("%-132s %s failed due to invalid params pointer.\n",
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

		logger.Printf("%-132s %s invoked sync method %s on client %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", methodId),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_INVOKE_ASYNC_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to invalid client handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}
		if paramsPtr == 0 {
			logger.Printf("%-132s %s failed due to invalid params pointer.\n",
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

		logger.Printf("%-132s %s invoked async method %s on client %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Yellow.Sprintf("0x%X", methodId),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_CONNECT:
		if client == nil {
			logger.Printf("%-132s %s failed due to invalid client handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}
		if _, err := GlobalFilesystem.Write(client.Name, make([]byte, FileBlockSize)); err != nil {
			logger.Printf("%-132s %s failed creating shared memory file (%s).\n",
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

		logger.Printf("%-132s %s connected %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_DISCONNECT_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to invalid client handle %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}
		if err := GlobalFilesystem.Delete(client.Name); err != nil {
			logger.Printf("%-132s %s failed deleting shared memory file (%s).\n",
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

		logger.Printf("%-132s %s disconnected %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_POLL_EVENT_FLAG:
		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		logger.Printf("%-132s %s polled event flag on %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown operation %s (handle=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ipmimgr_call"),
		color.Yellow.Sprintf("0x%X", op),
		color.Yellow.Sprintf("0x%X", handle),
	)
	return SCE_KERNEL_ERROR_ENOTSUP
}
