//go:build linux

package patcher

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// EmulateTcbAccess emulates a TCB access instruction and advances the instruction pointer.
func EmulateTcbAccess(signalContext *sys_struct.SIGNAL_CONTEXT, rip uint64) bool {
	dstReg, displacement, instructionLen, err := DecodeTcbAccess(rip)
	if err != nil {
		logger.Printf("Failed to emulate TCB access at 0x%X (%v).\n", rip, err)
		return false
	}

	threadContext := asm.GetCurrentThreadContext()
	tcbBase := uintptr(unsafe.Pointer(threadContext))
	value := *(*uint64)(unsafe.Pointer(tcbBase + uintptr(displacement)))

	mappedReg := MapGapstoneRegister(dstReg)
	if mappedReg == -1 {
		logger.Printf("Failed to map gapstone register %d.\n", dstReg)
		return false
	}

	signalContext.SetRegister(mappedReg, uintptr(value))
	signalContext.SetRegister(sys_struct.REG_RIP, uintptr(rip+instructionLen))

	logger.Printf(
		"Emulated TCB access at %s (loaded %s to register).\n",
		color.Yellow.Sprintf("0x%X", rip),
		color.Yellow.Sprintf("0x%X", value),
	)
	return true
}
