package lib

import (
	"crypto/rand"
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gc"
	"github.com/gookit/color"
)

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd, request, mode uintptr) uintptr {
	return libKernel_sys_ioctl(fd, request, mode)
}

func libKernel_sys_ioctl(fd, request, argPtr uintptr) uintptr {
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		return ENOENT
	}

	switch request {
	case SCE_NET_IOCTL_INIT:
		logger.Printf("%-132s %s initialized socket.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return 0
	case SCE_RNG_IOCTL_GET_ENTROPY:
		size := (request >> 16) & 0x1FFF
		argSlice := unsafe.Slice((*byte)(unsafe.Pointer(argPtr)), size)
		if _, err := rand.Read(argSlice); err != nil {
			return SCE_KERNEL_ERROR_EINVAL
		}

		logger.Printf("%-132s %s wrote %s random bytes to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", size),
			color.Yellow.Sprintf("0x%X", argPtr),
		)
		return 0
	case SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS:
		address := GlobalGraphicsController.SubmitDoneAddress
		WriteAddress(argPtr, address)

		logger.Printf("%-132s %s returned submit done address %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", address),
		)
		return 0
	case SCE_GC_IOCTL_GET_VM_ID:
		argSlice := unsafe.Slice((*byte)(unsafe.Pointer(argPtr)), 4)
		binary.LittleEndian.PutUint32(argSlice, 1)

		logger.Printf("%-132s %s returned vm id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return 0
	case SCE_GC_IOCTL_SET_RING_SIZES:
		ringSizes := (*GnmRingSizes)(unsafe.Pointer(argPtr))
		ring1Size := ringSizes.Ring1 * 256
		ring2Size := ringSizes.Ring2 * 256

		logger.Printf("%-132s %s tried setting ring sizes %s & %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", ring1Size),
			color.Yellow.Sprintf("0x%X", ring2Size),
		)
		return 0
	case SCE_GC_IOCTL_SET_MIP_STATS:
		logger.Printf("%-132s %s tried setting mip stats.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return 0
	case SCE_GC_IOCTL_GET_CU_MASK:
		computeUnitMask := (*GnmComputeUnitMask)(unsafe.Pointer(argPtr))
		computeUnitMask.Mask1 = 0xFFFFFFFF
		computeUnitMask.Mask2 = 0xFFFFFFFF
		computeUnitMask.Mask3 = 0
		computeUnitMask.Mask4 = 0

		logger.Printf("%-132s %s returned compute unit mask.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return 0
	case SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS:
		submitCommandBuffers := (*GnmSubmitCommandBuffers)(unsafe.Pointer(argPtr))

		logger.Printf("%-132s %s tried submitting command buffers (count=%s, flags=%s, buffersPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.ContextId),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.Count),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.IndirectBuffersPtr),
		)
		return 0
	case SCE_DCE_IOCTL_CMD:
		command := (*DceCommand)(unsafe.Pointer(argPtr))

		switch command.CommandId {
		case SCE_DCE_IOCTL_CMD_GET_CONNECTION_STATUS:
			logger.Printf("%-132s %s tried requesting connection status (handle=%s, param1=%s, param2=%s, param3=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", command.Handle),
				color.Yellow.Sprintf("0x%X", command.Param1),
				color.Yellow.Sprintf("0x%X", command.Param2),
				color.Yellow.Sprintf("0x%X", command.Param3),
			)
			return 0
		case SCE_DCE_IOCTL_CMD_GET_RESOLUTION_SUPPORT:
			logger.Printf("%-132s %s tried requesting resolution support (handle=%s, param1=%s, param2=%s, param3=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", command.Handle),
				color.Yellow.Sprintf("0x%X", command.Param1),
				color.Yellow.Sprintf("0x%X", command.Param2),
				color.Yellow.Sprintf("0x%X", command.Param3),
			)
			return 0
		case SCE_DCE_IOCTL_CMD_GET_ATTR_BUFFER_SIZE:
			size := GlobalDisplayCoreEngine.AttributeBufferSize
			if command.Param1 != 0 {
				// Attribute buffer offset.
				WriteAddress(command.Param1, 0)
			}
			if command.Param2 != 0 {
				// Attribute buffer size.
				WriteAddress(command.Param2, size)
			}

			logger.Printf("%-132s %s returned attribute buffer size %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", size),
			)
			return 0
		case SCE_DCE_IOCTL_CMD_GET_RESOLUTION_STATUS:
			if command.Param1 != 0 {
				resolutionInfo := (*DceResolutionStatus)(unsafe.Pointer(command.Param1))
				resolutionInfo.Width = 1920
				resolutionInfo.Height = 1080
				resolutionInfo.PaneWidth = 1920
				resolutionInfo.PaneHeight = 1080
				resolutionInfo.RefreshRate = SCE_DCE_REFRESH_RATE_59_94HZ
				resolutionInfo.ScreenSizeInches = 50
				resolutionInfo.Flags = 0
			}
			logger.Printf("%-132s %s returned resolution info (handle=%s, param1=%s, param2=%s, param3=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", command.Handle),
				color.Yellow.Sprintf("0x%X", command.Param1),
				color.Yellow.Sprintf("0x%X", command.Param2),
				color.Yellow.Sprintf("0x%X", command.Param3),
			)
			return 0
		case SCE_DCE_IOCTL_CMD_GET_PORT_STATUS_INFO:
			if command.Param1 != 0 {
				portStatus := (*DcePortStatusInfo)(unsafe.Pointer(command.Param1))
				portStatus.Connected = 1
			}
			logger.Printf("%-132s %s returned port status (handle=%s, param1=%s, param2=%s, param3=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", command.Handle),
				color.Yellow.Sprintf("0x%X", command.Param1),
				color.Yellow.Sprintf("0x%X", command.Param2),
				color.Yellow.Sprintf("0x%X", command.Param3),
			)
			return 0
		case SCE_DCE_IOCTL_CMD_SET_ATTR_BUFFER_ADDRESS:
			GlobalDisplayCoreEngine.AttributeBufferAddress = command.Param2

			logger.Printf("%-132s %s set attribute buffer address to %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", GlobalDisplayCoreEngine.AttributeBufferAddress),
			)
			return 0
		}

		logger.Printf("%-132s %s sent dce command %s (handle=%s, param1=%s, param2=%s, param3=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", command.CommandId),
			color.Yellow.Sprintf("0x%X", command.Handle),
			color.Yellow.Sprintf("0x%X", command.Param1),
			color.Yellow.Sprintf("0x%X", command.Param2),
			color.Yellow.Sprintf("0x%X", command.Param3),
		)
		return 0
	case SCE_DCE_IOCTL_REGISTER_BUFFERS:
		registerBuffers := (*DceRegisterBuffers)(unsafe.Pointer(argPtr))

		logger.Printf("%-132s %s tried registering buffers (commandId=%s, handle=%s, index=%s, address=%s, size=%s, flags=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", registerBuffers.CommandId),
			color.Yellow.Sprintf("0x%X", registerBuffers.Handle),
			color.Yellow.Sprintf("0x%X", registerBuffers.Index),
			color.Yellow.Sprintf("0x%X", registerBuffers.Address),
			color.Yellow.Sprintf("0x%X", registerBuffers.Size),
			color.Yellow.Sprintf("0x%X", registerBuffers.Flags),
		)
		return 0
	}

	logger.Printf("%-132s %s requested %s with argument at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[ioctl on %s]", file.Path),
		color.Yellow.Sprintf("0x%X", request),
		color.Yellow.Sprintf("0x%X", argPtr),
	)
	return 0
}
