//go:build windows

package sys_struct

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	CONTEXT_I386            = 0x00010000
	CONTEXT_CONTROL         = CONTEXT_I386 | 0x00000001
	CONTEXT_INTEGER         = CONTEXT_I386 | 0x00000002
	CONTEXT_SEGMENTS        = CONTEXT_I386 | 0x00000004
	CONTEXT_DEBUG_REGISTERS = CONTEXT_I386 | 0x00000010
	CONTEXT_FULL            = CONTEXT_CONTROL | CONTEXT_INTEGER | CONTEXT_SEGMENTS
)

// Hardware breakpoint types.
const (
	HW_EXECUTE   = 0
	HW_WRITE     = 1
	HW_READWRITE = 3
)

// Hardware breakpoint sizes.
const (
	HW_SIZE_1 = 0
	HW_SIZE_2 = 1
	HW_SIZE_4 = 3
	HW_SIZE_8 = 2
)

func SetHardwareBreakpoint(slot int, address uintptr, breakType, size int) error {
	if slot < 0 || slot > 3 {
		return syscall.EINVAL
	}

	// Get thread context with debug registers.
	var ctx SIGNAL_CONTEXT
	ctx.ContextFlags = CONTEXT_DEBUG_REGISTERS

	ret, _, err := GetThreadContext.Call(uintptr(windows.CurrentThread()), uintptr(unsafe.Pointer(&ctx)))
	if ret == 0 {
		return err
	}

	// Set the debug register (DR0-DR3).
	switch slot {
	case 0:
		ctx.Dr0 = uint64(address)
		break
	case 1:
		ctx.Dr1 = uint64(address)
		break
	case 2:
		ctx.Dr2 = uint64(address)
		break
	case 3:
		ctx.Dr3 = uint64(address)
		break
	}

	// Enable the breakpoint in DR7
	// DR7 format:
	// L0-L3: Local breakpoint enable (bits 0,2,4,6)
	// G0-G3: Global breakpoint enable (bits 1,3,5,7)
	// R/W0-R/W3: Read/Write flags (bits 16-17, 20-21, 24-25, 28-29)
	// LEN0-LEN3: Length flags (bits 18-19, 22-23, 26-27, 30-31)

	// Enable local breakpoint for this slot.
	ctx.Dr7 |= uint64(1 << (slot * 2))
	ctx.Dr7 |= uint64(1 << (slot*2 + 1))

	// Set type and size.
	rwShift := 16 + (slot * 4)
	lenShift := 18 + (slot * 4)

	// Clear existing bits.
	ctx.Dr7 &^= uint64(3 << rwShift)  // Clear R/W bits
	ctx.Dr7 &^= uint64(3 << lenShift) // Clear LEN bits

	// Set new bits.
	ctx.Dr7 |= uint64(breakType << rwShift)
	ctx.Dr7 |= uint64(size << lenShift)

	// Set the context back.
	ctx.ContextFlags = CONTEXT_DEBUG_REGISTERS
	ret, _, err = SetThreadContext.Call(uintptr(windows.CurrentThread()), uintptr(unsafe.Pointer(&ctx)))
	if ret == 0 {
		return err
	}

	return nil
}
