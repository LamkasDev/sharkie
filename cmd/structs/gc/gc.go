// Package gc contains structs to emulate the Graphics Controller (/dev/gc device).
package gc

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	"github.com/gookit/color"
)

var GlobalGraphicsController *GraphicsController

const (
	SCE_GC_IOCTL_SET_MIP_STATS           = 0xC0848119
	SCE_GC_IOCTL_GET_VM_ID               = 0xC004811F
	SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS = 0xC008811B
)

// GraphicsController keeps state of the /dev/gc device.
type GraphicsController struct {
	SubmitDoneAddress uintptr
	ActiveRingSlot    uint32
	PendingSubmits    uint32
	RingActive        bool
}

func NewGraphicsController() *GraphicsController {
	return &GraphicsController{
		SubmitDoneAddress: GlobalGoAllocator.Malloc(8),
	}
}

func (gc *GraphicsController) Read(b []byte) (int, error) {
	return 0, errors.New("gc read not implemented")
}

func (gc *GraphicsController) Write(b []byte) (int, error) {
	return 0, errors.New("gc write not implemented")
}

func (gc *GraphicsController) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("gc seek not implemented")
}

func (gc *GraphicsController) Close() error {
	return nil
}

func (gc *GraphicsController) Stat() (fs.FileInfo, error) {
	return nil, errors.New("gc stat not implemented")
}

func (gc *GraphicsController) Truncate(size int64) error {
	return errors.New("gc truncate not implemented")
}

func (gc *GraphicsController) Ioctl(request uint32, argPtr uintptr) error {
	switch request {
	case SCE_GC_IOCTL_SET_MIP_STATS:
		logger.Printf("%-132s %s tried setting mip stats.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return nil
	case SCE_GC_IOCTL_GET_VM_ID:
		argSlice := unsafe.Slice((*byte)(unsafe.Pointer(argPtr)), 4)
		binary.LittleEndian.PutUint32(argSlice, 1)

		logger.Printf("%-132s %s returned vm id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return nil
	case SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS:
		address := GlobalGraphicsController.SubmitDoneAddress
		WriteAddress(argPtr, address)

		logger.Printf("%-132s %s returned submit done address %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", address),
		)
		return nil
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
		return nil
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
		return nil
	case SCE_GC_IOCTL_SUBMIT_DONE:
		GlobalGraphicsController.RingActive = false
		GlobalGraphicsController.PendingSubmits = 0
		WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))

		logger.Printf("%-132s %s flushed ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
		return nil
	case SCE_GC_IOCTL_DRAIN_RING:
		logger.Printf("%-132s %s tried draining ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
		return nil
	case SCE_GC_IOCTL_SET_WORK_MODE:
		setWorkMode := (*GnmSetWorkMode)(unsafe.Pointer(argPtr))
		logger.Printf("%-132s %s tried setting work mode to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", setWorkMode.Mode),
		)
		return nil
	case SCE_GC_IOCTL_SWITCH_BUFFER:
		switchBuffer := (*GnmSwitchBuffer)(unsafe.Pointer(argPtr))
		GlobalGraphicsController.ActiveRingSlot = switchBuffer.RingSlot

		logger.Printf("%-132s %s switched to ring slot %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", switchBuffer.RingSlot),
		)
		return nil
	case SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS:
		submitBuffers := (*GnmSubmitCommandBuffers)(unsafe.Pointer(argPtr))
		if submitBuffers.Count == 0 || submitBuffers.IndirectBuffersPtr == 0 {
			logger.Printf("%-132s %s skipped due to zero size or invalid pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl[gc:submit_cbs]"),
			)
			return nil
		}

		buffers := unsafe.Slice((*PM4IndirectBuffer)(unsafe.Pointer(submitBuffers.IndirectBuffersPtr)), submitBuffers.Count)
		buffersCopy := make([]PM4IndirectBuffer, submitBuffers.Count)
		copy(buffersCopy, buffers)

		GlobalLiverpool.SubmitCommandBuffers(buffersCopy)
		WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))
		GlobalGraphicsController.PendingSubmits += submitBuffers.Count

		logger.Printf("%-132s %s submitted %s indirect buffers (count=%s, contextId=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Green.Sprintf("%d", len(buffers)),
			color.Green.Sprintf("%d", submitBuffers.Count),
			color.Yellow.Sprintf("0x%X", submitBuffers.ContextId),
		)
		return nil
	case SCE_GC_IOCTL_DINGDONG:
		dingDong := (*GnmDingDong)(unsafe.Pointer(argPtr))
		GlobalGraphicsController.PendingSubmits++

		logger.Printf("%-132s %s dinged (pipe=%s, queue=%s, slot=%s, wp=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("%d", dingDong.PipeIndex),
			color.Yellow.Sprintf("%d", dingDong.QueueIndex),
			color.Yellow.Sprintf("%d", dingDong.SlotIndex),
			color.Yellow.Sprintf("0x%X", dingDong.WritePointer),
		)
		return nil
	}

	return errors.New("unknown gc ioctl")
}

func SetupGraphicsController() {
	GlobalGraphicsController = NewGraphicsController()
}
