package asm

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

var (
	// Global constants.
	StubAddr             uintptr
	ExceptionHandlerAddr uintptr

	// TLS configuration.
	GoTlsSlot   uintptr
	GoTlsOffset uintptr

	PlaystationTlsSlot   uintptr
	PlaystationTlsOffset uintptr
)

type ThreadContext struct {
	ThreadId uintptr

	// Stack switching.
	SystemSP      uintptr
	PlaystationSP uintptr
	GoSP          uintptr
	LastGoSP      uintptr
	GoBP          uintptr
	SavedG        uintptr

	// Execution state.
	ReturnAddressAnchor uintptr
	GlobalStubContext   uintptr
	GlobalExceptionInfo uintptr

	// Saved state for Call.
	CallSavedBP  uintptr
	CallSavedBX  uintptr
	CallSavedR12 uintptr
	CallSavedR13 uintptr
	CallSavedR14 uintptr
	CallSavedR15 uintptr
	CallSavedSP  uintptr
}

func NewThreadContext(threadId int32, stackPtr uintptr) *ThreadContext {
	return &ThreadContext{
		ThreadId:      uintptr(threadId),
		PlaystationSP: stackPtr,
	}
}

func GetCurrentThreadContext() *ThreadContext

func SetThreadContext(ctx *ThreadContext) {
	sys_struct.SetTlsSlot(GoTlsSlot, uintptr(unsafe.Pointer(ctx)))
}

func AllocTlsSlots() {
	GoTlsSlot, GoTlsOffset = sys_struct.AllocTlsSlot()
	PlaystationTlsSlot, PlaystationTlsOffset = sys_struct.AllocTlsSlot()
	if GoTlsSlot >= 64 || PlaystationTlsSlot >= 64 {
		panic("tls slot is too high, this is not supported")
	}
}
