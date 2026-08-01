package gnm_driver

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/irq"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000002280
// __int64 __fastcall sceGnmAddEqEvent(__int64, int, __int64)
func libSceGnmDriver_sceGnmAddEqEvent(equeueHandle, id, userData uintptr) uintptr {
	equeue := GetEqueue(equeueHandle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to invalid equeue.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmAddEqEvent"),
		)
		return 0x80020009
	}

	GlobalInterruptHandler.Register(InterruptGraphicsFlip, func(irqType InterruptType) {
		event := KernelEvent{
			Id:          uint64(id),
			Filter:      EVFILT_GRAPHICS_CORE,
			Flags:       EV_ADD,
			FilterFlags: 0,
			FilterData:  uint64(id),
			UserData:    userData,
		}

		select {
		case equeue.Events <- event:
		default:
			logger.Printf("equeue event channel full.\n")
		}
	})

	logger.Printf("%-132s %s added flip event to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceGnmAddEqEvent"),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}
