package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

const (
	CTL_SYSCTL = 0
	CTL_KERN   = 1
	CTL_HW     = 6
)

const (
	RTP_LOOKUP        = 0
	RTP_PRIO_NORMAL   = 1
	RTP_PRIO_REALTIME = 2
)

const (
	UMTX_OP_WAIT              = 2
	UMTX_OP_WAKE              = 3
	UMTX_OP_WAIT_UINT_PRIVATE = 15
	UMTX_OP_WAKE_PRIVATE      = 16
)

const (
	REGMGR_GET_INT = 2
	REGMGR_SET_INT = 6
	REGMGR_GET_BIN = 25
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
		fmt.Printf("%-120s %s failed due to invalid name pointer.\n",
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
		fmt.Printf("%-120s %s failed due to unknown OIDs %+v.\n",
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
			fmt.Printf("%-120s %s failed due to invalid argument pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_sysarch"),
			)
			SetErrno(EFAULT)
			return ERR_PTR
		}

		argsPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(argsPtr)), 8)
		tcbBaseAddr := uintptr(binary.LittleEndian.Uint64(argsPtrSlice))
		emu.GlobalModuleManager.Tcb = (*Tcb)(unsafe.Pointer(tcbBaseAddr))

		ret, _, _ := sys_struct.TlsSetValue.Call(sys_struct.TlsSlot, tcbBaseAddr)
		if ret == 0 {
			fmt.Printf("%-120s %s failed setting TCB base.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_sysarch"),
			)
			SetErrno(EPERM)
			return ERR_PTR
		}

		fmt.Printf(
			"%-120s %s set TCB base to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_sysarch"),
			color.Yellow.Sprintf("0x%X", tcbBaseAddr),
		)
		return 0
	}

	fmt.Printf("%-120s %s failed due to unknown number %s.\n",
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
		fmt.Printf("%-120s %s failed due to invalid Id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_thr_self"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	thread := emu.GlobalModuleManager.Tcb.Thread
	idSlice := unsafe.Slice((*byte)(unsafe.Pointer(idPtr)), 8)
	binary.LittleEndian.PutUint64(idSlice, uint64(thread.ThreadId))

	fmt.Printf("%-120s %s requested thread id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_thr_self"),
		color.Green.Sprintf("%d", thread.ThreadId),
	)
	return 0
}

// 0x0000000000001710
// __int64 __fastcall rtprio_thread()
func libKernel_rtprio_thread(function, lwpid, rtpPtr uintptr) uintptr {
	if rtpPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid structs pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("rtprio_thread"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	if function != RTP_LOOKUP {
		fmt.Printf("%-120s %s failed due to unknown function %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("rtprio_thread"),
			color.Yellow.Sprintf("0x%X", function),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	rtpSlice := unsafe.Slice((*byte)(unsafe.Pointer(rtpPtr)), 4)
	binary.LittleEndian.PutUint16(rtpSlice, RTP_PRIO_NORMAL)
	binary.LittleEndian.PutUint16(rtpSlice[2:], 0)

	fmt.Printf("%-120s %s requested rtp struct (type=%s, priority=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("rtprio_thread"),
		color.Yellow.Sprintf("0x%X", RTP_PRIO_NORMAL),
		color.Yellow.Sprintf("0x%X", 0),
	)
	return 0
}

// 0x0000000000002BA0
// __int64 sub_2BA0()
func libKernel_sys_umtx_op(objPtr, op, val, uaddr, uaddr2 uintptr) uintptr {
	switch op {
	case UMTX_OP_WAKE, UMTX_OP_WAKE_PRIVATE:
		fmt.Printf("%-120s %s tried waking up thread xd.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_umtx_op"),
		)
		return 0
	case UMTX_OP_WAIT, UMTX_OP_WAIT_UINT_PRIVATE:
		objSlice := unsafe.Slice((*byte)(unsafe.Pointer(objPtr)), 4)
		obj := uintptr(binary.LittleEndian.Uint32(objSlice))
		if obj != val {
			fmt.Printf("%-120s %s skipped wait because %s != %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_umtx_op"),
				color.Yellow.Sprintf("0x%X", obj),
				color.Yellow.Sprintf("0x%X", val),
			)
			return 0
		}

		fmt.Printf("%-120s %s waiting on %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_umtx_op"),
			color.Yellow.Sprintf("0x%X", objPtr),
		)
		return 0
	}

	fmt.Printf("%-120s %s failed due to unknown operation %s.\n",
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
		fmt.Printf("%-120s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_get_authinfo"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	if processId != 0 && processId != 1001 {
		fmt.Printf("%-120s %s is requesting invalid process id %s.\n",
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

	fmt.Printf("%-120s %s returning auth info for process id %s (infoPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_get_authinfo"),
		color.Green.Sprintf("%d", processId),
		color.Yellow.Sprintf("0x%X", infoPtr),
	)
	return 0
}

// 0x00000000000017F0
// __int64 __fastcall _sys_regmgr_call()
func libKernel___sys_regmgr_call(op, id, resultPtr, valuePtr, size uintptr) uintptr {
	switch op {
	case REGMGR_GET_INT:
		if valuePtr == 0 || size < 4 {
			fmt.Printf("%-120s %s failed due to invalid value pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__sys_regmgr_call"),
			)
			return EFAULT
		}
		valueSlice := unsafe.Slice((*byte)(unsafe.Pointer(valuePtr)), size)
		for i := 0; i < len(valueSlice); i += 4 {
			binary.LittleEndian.PutUint32(valueSlice[i:], 0)
		}

		fmt.Printf("%-120s %s requested integer (id=%s, resultPtr=%s, valuePtr=%s, size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
			color.Yellow.Sprintf("0x%X", id),
			color.Yellow.Sprintf("0x%X", resultPtr),
			color.Yellow.Sprintf("0x%X", valuePtr),
			color.Green.Sprintf("%d", size),
		)
		return 0
	}

	fmt.Printf("%-120s %s failed due to unknown operation %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_regmgr_call"),
		color.Green.Sprintf("%d", op),
	)
	return ENOENT
}

// 0x0000000000001F10
// __int64 __fastcall _sys_get_proc_type_info()
func libKernel___sys_get_proc_type_info(infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid info pointer.\n",
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

	fmt.Printf("%-120s %s returning process type info (infoPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_get_proc_type_info"),
		color.Yellow.Sprintf("0x%X", infoPtr),
	)
	return 0
}

// 0x00000000000289C0
// __int64 __fastcall _tls_get_addr(_QWORD *, __int64, __int64, __int64, __int64, int)
func libKernel___tls_get_addr(tlsIndexPtr uintptr) uintptr {
	if tlsIndexPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid tls index pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
		)
		return EFAULT
	}

	tlsIndex := (*TlsIndex)(unsafe.Pointer(tlsIndexPtr))
	address, ok := TlsBaseRepo[tlsIndex.ModuleId]
	if !ok {
		fmt.Printf("%-120s %s failed due to invalid module index %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
			color.Green.Sprint(tlsIndex.ModuleId),
		)
		return 0
	}

	fmt.Printf("%-120s %s returning tls address %s for module %s (offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__tls_get_addr"),
		color.Yellow.Sprintf("0x%X", address),
		color.Green.Sprintf("%d", tlsIndex.ModuleId),
		color.Yellow.Sprintf("0x%X", tlsIndex.Offset),
	)
	return address + tlsIndex.Offset
}
