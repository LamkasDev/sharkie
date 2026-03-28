package asm

import (
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

var (
	// ThreadRepo maps thread IDs to host thread contexts.
	ThreadContextRepo   = map[int32]*ThreadContext{}
	ThreadContextPinner = runtime.Pinner{}

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
	ReturnAddressAnchor       uintptr
	GlobalStubContext         uintptr
	GlobalExceptionInfo       uintptr    // For Windows.
	GlobalExceptionInfoBuffer [2]uintptr // For Linux.

	// Saved state for Call.
	CallSavedBX  uintptr
	CallSavedR12 uintptr
	CallSavedR13 uintptr
	CallSavedR14 uintptr
	CallSavedR15 uintptr
}

func NewThreadContext(threadId int32, stackPtr uintptr) *ThreadContext {
	threadContext := &ThreadContext{
		ThreadId:      uintptr(threadId),
		PlaystationSP: stackPtr,
	}
	ThreadContextRepo[threadId] = threadContext
	ThreadContextPinner.Pin(threadContext)

	return threadContext
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
