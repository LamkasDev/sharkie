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
			logger.Printf("%-132s %s failed due to unknown client %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}

		delete(GlobalIpmiManager.Clients, uint32(handle))
		delete(GlobalIpmiManager.Servers, uint32(handle))
		if resultPtr != 0 {
			WriteResult(resultPtr, 0)
		}

		logger.Printf("%-132s %s destroyed ipmi client %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_INVOKE_SYNC_CLIENT, IMPI_INVOKE_ASYNC_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to unknown client %s.\n",
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

		return InvokeImpiClientMethod(client, resultPtr, paramsPtr, paramsSize, objPtr)

	case IMPI_CONNECT:
		if client == nil {
			logger.Printf("%-132s %s failed due to unknown client %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Yellow.Sprintf("0x%X", handle),
			)
			return SCE_KERNEL_ERROR_ENOENT
		}
		server = GetImpiServerByName(client.Name)
		if server == nil {
			logger.Printf("%-132s %s failed due to unknown server %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Blue.Sprint(client.Name),
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
			WriteResult(resultPtr, 0)
		}

		logger.Printf("%-132s %s connected %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ipmimgr_call"),
			color.Blue.Sprint(client.Name),
		)
		return 0

	case IMPI_DISCONNECT_CLIENT:
		if client == nil {
			logger.Printf("%-132s %s failed due to unknown client %s.\n",
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
			WriteResult(resultPtr, 0)
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
		if client == nil {
			logger.Printf("%-132s %s failed due to unknown client %s.\n",
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

		client.EventFlag.Lock.Lock()
		defer client.EventFlag.Lock.Unlock()

		pollEventFlag := (*IpmiPollEventFlag)(unsafe.Pointer(paramsPtr))
		if CheckEventFlagCondition(client.EventFlag.CurrentPattern, uint64(pollEventFlag.WaitPattern), pollEventFlag.WaitMode) {
			if pollEventFlag.OutPatternPtr != 0 {
				outPatternSlice := unsafe.Slice((*byte)(unsafe.Pointer(pollEventFlag.OutPatternPtr)), 8)
				binary.LittleEndian.PutUint64(outPatternSlice, client.EventFlag.CurrentPattern)
			}

			if (pollEventFlag.WaitMode & EVF_WAITMODE_CLEAR_ALL) != 0 {
				client.EventFlag.CurrentPattern = 0
			}
			if (pollEventFlag.WaitMode & EVF_WAITMODE_CLEAR_PAT) != 0 {
				client.EventFlag.CurrentPattern &= ^uint64(pollEventFlag.WaitPattern)
			}

			if resultPtr != 0 {
				WriteResult(resultPtr, 0)
			}

			logger.Printf("%-132s %s finished waiting on event flag %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ipmimgr_call"),
				color.Blue.Sprint(client.EventFlag.Name),
			)
			return 0
		}

		logger.Printf("%-132s %s tried waiting on event flag %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_trywait"),
			color.Blue.Sprint(client.EventFlag.Name),
		)
		return SCE_KERNEL_ERROR_TIMEDOUT
	}

	logger.Printf("%-132s %s failed due to unknown operation %s (handle=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ipmimgr_call"),
		color.Yellow.Sprintf("0x%X", op),
		color.Yellow.Sprintf("0x%X", handle),
	)
	return SCE_KERNEL_ERROR_ENOTSUP
}

func InvokeImpiClientMethod(client *IpmiClient, resultPtr, paramsPtr, paramsSize, objPtr uintptr) uintptr {
	paramsSlice := unsafe.Slice((*byte)(unsafe.Pointer(paramsPtr)), paramsSize)
	methodId := binary.LittleEndian.Uint32(paramsSlice)
	inputSize := binary.LittleEndian.Uint32(paramsSlice[0x4:])
	outputSize := binary.LittleEndian.Uint32(paramsSlice[0x8:])
	inputPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x10:]))
	outputPtr := uintptr(binary.LittleEndian.Uint64(paramsSlice[0x18:]))

	switch client.Name {
	case "SceNpMgrIpc":
		switch methodId {
		case IMPI_METHOD_SERVICE_INIT:
			if inputSize >= 4 && inputPtr != 0 {
				inputSlice := unsafe.Slice((*byte)(unsafe.Pointer(inputPtr)), inputSize)
				serviceId := binary.LittleEndian.Uint32(inputSlice)

				if resultPtr != 0 {
					WriteResult(resultPtr, 0)
				}

				logger.Printf("%-132s %s initialized service %s on client %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("ipmimgr_call"),
					color.Yellow.Sprintf("0x%X", serviceId),
					color.Blue.Sprint(client.Name),
				)
				return 0
			}

			if outputSize > 0 && outputPtr != 0 {
				outputSlice := unsafe.Slice((*byte)(unsafe.Pointer(outputPtr)), 8)
				namePtr := uintptr(binary.LittleEndian.Uint64(outputSlice))

				WriteCString(namePtr, client.Name)

				if resultPtr != 0 {
					WriteResult(resultPtr, 0)
				}

				logger.Printf("%-132s %s returned server name for client %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("ipmimgr_call"),
					color.Blue.Sprint(client.Name),
				)
				return 0
			}
			break
		}
		break
	}

	if resultPtr != 0 {
		WriteResult(resultPtr, 0)
	}

	logger.Printf("%-132s %s invoked unknown method %s on client %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ipmimgr_call"),
		color.Yellow.Sprintf("0x%X", methodId),
		color.Blue.Sprint(client.Name),
	)
	return 0
}
