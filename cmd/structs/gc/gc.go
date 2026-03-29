// Package gc contains structs to emulate the Graphics Controller (/dev/gc device).
package gc

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

var GlobalGraphicsController *GraphicsController

const (
	SCE_GC_IOCTL_GET_VM_ID               = 0xC004811F
	SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS = 0xC008811B
	SCE_GC_IOCTL_SET_MIP_STATS           = 0xC0848119
)

// GraphicsController keeps state of the /dev/gc device.
type GraphicsController struct {
	SubmitDoneAddress uintptr
	ActiveRingSlot    uint32
	PendingSubmits    uint32
}

func NewGraphicsController() *GraphicsController {
	return &GraphicsController{
		SubmitDoneAddress: structs.GlobalGoAllocator.Malloc(8),
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
	case SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS:
		address := GlobalGraphicsController.SubmitDoneAddress
		structs.WriteAddress(argPtr, address)

		logger.Printf("%-132s %s returned submit done address %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", address),
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
	case SCE_GC_IOCTL_SET_MIP_STATS:
		logger.Printf("%-132s %s tried setting mip stats.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
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
	case SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS:
		submitCommandBuffers := (*GnmSubmitCommandBuffers)(unsafe.Pointer(argPtr))

		logger.Printf("%-132s %s tried submitting command buffers (count=%s, flags=%s, buffersPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.ContextId),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.Count),
			color.Yellow.Sprintf("0x%X", submitCommandBuffers.IndirectBuffersPtr),
		)
		return nil
	}

	return errors.New("unknown gc ioctl")
}

func SetupGraphicsController() {
	GlobalGraphicsController = NewGraphicsController()
}
