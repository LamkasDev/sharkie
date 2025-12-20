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
		return ^uintptr(0)
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
		return ^uintptr(0)
	}
	if err != 0 {
		SetErrno(err)
		return ^uintptr(0)
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
			return ^uintptr(0)
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
			return ^uintptr(0)
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
	return ^uintptr(0)
}

// 0x0000000000001590
// __int64 __fastcall sub_1590()
func libKernel_sys_thr_self(idPtr uintptr) uintptr {
	if idPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid ID pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_thr_self"),
		)
		SetErrno(EFAULT)
		return ^uintptr(0)
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
		return ^uintptr(0)
	}
	if function != RTP_LOOKUP {
		fmt.Printf("%-120s %s failed due to unknown function %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("rtprio_thread"),
			color.Yellow.Sprintf("0x%X", function),
		)
		SetErrno(EINVAL)
		return ^uintptr(0)
	}

	rtpSlice := unsafe.Slice((*byte)(unsafe.Pointer(rtpPtr)), 4)
	binary.LittleEndian.PutUint16(rtpSlice, RTP_PRIO_NORMAL)
	binary.LittleEndian.PutUint16(rtpSlice[2:], 0)

	fmt.Printf("%-120s %s requested rtp structs (type=%s, priority=%s).\n",
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
