package posix

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/mutex"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func libScePosix_sys_umtx_op(objPtr, op, value, uaddr, uaddr2 uintptr) uintptr {
	if objPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid object pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_umtx_op"),
		)
		return EINVAL
	}
	userMutex := GetUserMutex(objPtr)

	switch op {
	case UMTX_OP_WAKE, UMTX_OP_WAKE_PRIVATE:
		logger.Printf("%-132s %s waking up %s (value=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_umtx_op"),
			color.Yellow.Sprintf("0x%X", objPtr),
			color.Yellow.Sprintf("0x%X", value),
		)
		if value == 1 {
			userMutex.Signal()
		} else {
			userMutex.Broadcast()
		}
		return 0
	case UMTX_OP_WAIT_UINT, UMTX_OP_WAIT_UINT_PRIVATE:
		userMutex.L.Lock()
		objSlice := unsafe.Slice((*byte)(unsafe.Pointer(objPtr)), 4)
		obj := uintptr(binary.LittleEndian.Uint32(objSlice))
		if obj != value {
			userMutex.L.Unlock()
			logger.Printf("%-132s %s skipped wait because %s != %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_umtx_op"),
				color.Yellow.Sprintf("0x%X", obj),
				color.Yellow.Sprintf("0x%X", value),
			)
			return 0
		}

		// TODO: implement timeout.
		logger.Printf("%-132s %s waiting on %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_umtx_op"),
			color.Yellow.Sprintf("0x%X", objPtr),
		)
		userMutex.Wait()
		userMutex.L.Unlock()
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown operation %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_umtx_op"),
		color.Yellow.Sprintf("0x%X", op),
	)
	return EINVAL
}
