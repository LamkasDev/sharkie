package asm

import (
	"sync"
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
)

type ThreadContext struct {
	ThreadId uintptr

	// Stack switching.
	WindowsSP     uintptr
	PlaystationSP uintptr
	GoSP          uintptr
	GoBP          uintptr
	SavedG        uintptr

	// Execution state.
	ReturnAddressAnchor uintptr
	CallReturnAddress   uintptr
	GlobalStubContext   uintptr
	GlobalExceptionInfo uintptr
}

func AllocGoTlsSlot() {
	slot, _, err := sys_struct.TlsAlloc.Call()
	if slot == 0 {
		panic(err)
	}
	GoTlsSlot = slot
	GoTlsOffset = 0x1480 + slot*8
}

var ThreadContexts = make(map[int32]*ThreadContext)
var ThreadContextLock sync.Mutex

func NewThreadContext(threadId int32, stackPtr uintptr) *ThreadContext {
	threadContext := &ThreadContext{
		ThreadId:      uintptr(threadId),
		PlaystationSP: stackPtr,
	}
	ThreadContextLock.Lock()
	ThreadContexts[threadId] = threadContext
	ThreadContextLock.Unlock()

	return threadContext
}

func SetThreadContext(ctx *ThreadContext) {
	status, _, err := sys_struct.TlsSetValue.Call(GoTlsSlot, uintptr(unsafe.Pointer(ctx)))
	if status == 0 {
		panic(err)
	}
}

func GetCurrentThreadContext() *ThreadContext
