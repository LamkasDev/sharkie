package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/mutex"
	. "github.com/LamkasDev/sharkie/cmd/structs/tcb"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

const (
	CTL_SYSCTL = 0
	CTL_KERN   = 1
	CTL_HW     = 6
)

const (
	UMTX_OP_WAIT              = 2
	UMTX_OP_WAKE              = 3
	UMTX_OP_WAIT_UINT         = 11
	UMTX_OP_WAIT_UINT_PRIVATE = 15
	UMTX_OP_WAKE_PRIVATE      = 16
)

const AMD64_SET_FSBASE = 129

type RtPriority struct {
	Type     uint16
	Priority uint16
}

// 0x00000000000111F0
// __int64 __fastcall sysctl(_DWORD *, int, _DWORD *_RDX, unsigned __int64 *, __int64)
func libKernel_sysctl(namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) uintptr {
	// Perform initial checks.
	if namePtr == 0 || nameLen < 2 {
		logger.Printf("%-132s %s failed due to invalid name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	// Resolve MIBs, fancy name for string oooooofsadfasv.
	mib := make([]uint32, nameLen)
	for i := uint32(0); i < nameLen; i++ {
		mibSlice := unsafe.Slice((*byte)(unsafe.Pointer(namePtr+uintptr(i*4))), 4)
		mib[i] = binary.LittleEndian.Uint32(mibSlice)
	}

	err, found := uintptr(0), false
	switch mib[0] {
	case CTL_SYSCTL:
		err, found = libKernel_ctl_sysctl(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen)
		break
	case CTL_KERN:
		err, found = libKernel_ctl_kern(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen)
		break
	case CTL_HW:
		err, found = libKernel_ctl_hw(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen)
		break
	}
	if !found {
		logger.Printf("%-132s %s failed due to unknown OIDs %+v.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
			mib,
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	if err != 0 {
		SetErrno(err)
		return ERR_PTR
	}

	return err
}

// 0x0000000000000F70
// __int64 __fastcall sysarch()
func libKernel_sys_sysarch(number uintptr, argsPtr uintptr) uintptr {
	switch number {
	case AMD64_SET_FSBASE:
		if argsPtr == 0 {
			logger.Printf("%-132s %s failed due to invalid argument pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_sysarch"),
			)
			SetErrno(EFAULT)
			return ERR_PTR
		}

		thread := emu.GetCurrentThread()
		argsPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(argsPtr)), 8)
		tcbBaseAddr := uintptr(binary.LittleEndian.Uint64(argsPtrSlice))
		thread.Tcb = (*Tcb)(unsafe.Pointer(tcbBaseAddr))
		sys_struct.SetTlsSlot(asm.PlaystationTlsSlot, tcbBaseAddr)

		logger.Printf(
			"%-132s %s set TCB base to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_sysarch"),
			color.Yellow.Sprintf("0x%X", tcbBaseAddr),
		)
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown number %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_sysarch"),
		color.Yellow.Sprintf("0x%X", number),
	)
	SetErrno(EINVAL)
	return ERR_PTR
}

// 0x0000000000001590
// __int64 __fastcall sub_1590()
func libKernel_sys_thr_self(idPtr uintptr) uintptr {
	if idPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_thr_self"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	thread := emu.GetCurrentThread()
	idSlice := unsafe.Slice((*byte)(unsafe.Pointer(idPtr)), 8)
	binary.LittleEndian.PutUint64(idSlice, uint64(thread.Id))

	logger.Printf("%-132s %s requested thread id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_thr_self"),
		color.Green.Sprintf("%d", thread.Id),
	)
	return 0
}

// 0x0000000000002BA0
// __int64 sub_2BA0()
func libKernel_sys_umtx_op(objPtr, op, value, uaddr, uaddr2 uintptr) uintptr {
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

// 0x0000000000001C70
// __int64 __fastcall get_authinfo()
func libKernel_sys_get_authinfo(processId uintptr, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_get_authinfo"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	if processId != 0 && processId != 1001 {
		logger.Printf("%-132s %s is requesting invalid process id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_get_authinfo"),
			color.Yellow.Sprintf("0x%X", processId),
		)
	}

	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 136)
	for i := 0; i < len(infoSlice); i += 8 {
		binary.LittleEndian.PutUint64(infoSlice[i:], 0)
	}
	binary.LittleEndian.PutUint64(infoSlice[8:], 0x6000000000000000)

	logger.Printf("%-132s %s returning auth info for process id %s (infoPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_get_authinfo"),
		color.Green.Sprintf("%d", processId),
		color.Yellow.Sprintf("0x%X", infoPtr),
	)
	return 0
}

// 0x0000000000001F10
// __int64 __fastcall _sys_get_proc_type_info()
func libKernel___sys_get_proc_type_info(infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_get_proc_type_info"),
		)
		return EFAULT
	}

	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 12)
	// size := uintptr(binary.LittleEndian.Uint32(infoSlice))
	flags := uint32(0)
	binary.LittleEndian.PutUint32(infoSlice[4:], PROC_TYPE_BIG_APP)
	binary.LittleEndian.PutUint32(infoSlice[8:], flags)

	logger.Printf("%-132s %s returning process type info (infoPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_get_proc_type_info"),
		color.Yellow.Sprintf("0x%X", infoPtr),
	)
	return 0
}
