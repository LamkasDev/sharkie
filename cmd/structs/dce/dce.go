// Package dce contains structs to emulate the Display Core Engine (/dev/dce device).
package dce

import (
	"errors"
	"io/fs"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/gookit/color"
)

var GlobalDisplayCoreEngine *DisplayCoreEngine

// DisplayCoreEngine keeps state of the /dev/dce device.
type DisplayCoreEngine struct {
	Handles                [VideoOutMaxHandles]VideoOutHandle
	AttributeBufferAddress uintptr
	AttributeBufferSize    uintptr
}

func NewDisplayCoreEngine() *DisplayCoreEngine {
	dce := &DisplayCoreEngine{
		AttributeBufferSize: 0x4000,
	}
	for i := range dce.Handles {
		dce.Handles[i].Id = i + 1
		dce.Handles[i].LabelBufferAddress = GlobalGoAllocator.Malloc(uintptr(VideoOutMaxBuffers) * 8)
	}

	return dce
}

func (dce *DisplayCoreEngine) Read(b []byte) (int, error) {
	return 0, errors.New("dce read not implemented")
}

func (dce *DisplayCoreEngine) Write(b []byte) (int, error) {
	return 0, errors.New("dce write not implemented")
}

func (dce *DisplayCoreEngine) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("dce seek not implemented")
}

func (dce *DisplayCoreEngine) Close() error {
	return nil
}

func (dce *DisplayCoreEngine) Stat() (fs.FileInfo, error) {
	return nil, errors.New("dce stat not implemented")
}

func (dce *DisplayCoreEngine) Truncate(size int64) error {
	return errors.New("dce truncate not implemented")
}

func (dce *DisplayCoreEngine) Ioctl(request uint32, argPtr uintptr) error {
	switch request {
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
			return nil
		case SCE_DCE_IOCTL_CMD_GET_RESOLUTION_SUPPORT:
			logger.Printf("%-132s %s tried requesting resolution support (handle=%s, param1=%s, param2=%s, param3=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", command.Handle),
				color.Yellow.Sprintf("0x%X", command.Param1),
				color.Yellow.Sprintf("0x%X", command.Param2),
				color.Yellow.Sprintf("0x%X", command.Param3),
			)
			return nil
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
			return nil
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
			return nil
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
			return nil
		case SCE_DCE_IOCTL_CMD_SET_ATTR_BUFFER_ADDRESS:
			GlobalDisplayCoreEngine.AttributeBufferAddress = command.Param2

			logger.Printf("%-132s %s set attribute buffer address to %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", GlobalDisplayCoreEngine.AttributeBufferAddress),
			)
			return nil
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
		return nil
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
		return nil
	}

	return errors.New("unknown dce ioctl")
}

func (dce *DisplayCoreEngine) GetHandleById(id int) *VideoOutHandle {
	if id < 1 || id > VideoOutMaxHandles {
		return nil
	}

	return &dce.Handles[id-1]
}

func SetupDisplayCoreEngine() {
	GlobalDisplayCoreEngine = NewDisplayCoreEngine()
}
