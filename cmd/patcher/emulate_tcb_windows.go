//go:build windows

package patcher

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// EmulateTcbAccess emulates a TCB access instruction and advances the instruction pointer.
func EmulateTcbAccess(ctx *sys_struct.CONTEXT, rip uint64) bool {
	dstReg, displacement, instructionLen, err := DecodeTcbAccess(rip)
	if err != nil {
		logger.Printf("Failed to emulate TCB access at 0x%X (%v).\n", rip, err)
		return false
	}

	threadContext := asm.GetCurrentThreadContext()
	tcbBase := uintptr(unsafe.Pointer(threadContext))
	value := *(*uint64)(unsafe.Pointer(tcbBase + uintptr(displacement)))

	SetContextRegister(ctx, dstReg, value)
	ctx.Rip = rip + instructionLen

	logger.Printf(
		"Emulated TCB access at %s (loaded %s to register).\n",
		color.Yellow.Sprintf("0x%X", rip),
		color.Yellow.Sprintf("0x%X", value),
	)
	return true
}
