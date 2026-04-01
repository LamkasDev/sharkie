package asm

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

var (
	// ThreadContextRepo maps thread pointers to host thread contexts.
	ThreadContextRepo = make(map[uintptr]*ThreadContext)

	// ThreadContextLock protects ThreadContextRepo, so multiple threads can look up thread contexts safely.
	ThreadContextLock sync.RWMutex

	// ThreadContextPinner prevents ThreadContext objects from being moved by the Go garbage collector.
	ThreadContextPinner = runtime.Pinner{}

	// Addresses of assembly functions.
	StubAddr             uintptr
	ExceptionHandlerAddr uintptr

	// TLS configuration for Go.
	GoTlsSlot   uintptr
	GoTlsOffset uintptr

	// TLS configuration for guest threads.
	PlaystationTlsSlot   uintptr
	PlaystationTlsOffset uintptr
)

// ThreadContext holds thread-local execution state.
type ThreadContext struct {
	ThreadPtr uintptr // Pointer to thread.

	// Stack switching related fields.
	SystemSP      uintptr // Stack pointer when in the host context.
	PlaystationSP uintptr // Stack pointer when executing guest code.
	GoSP          uintptr // Stack pointer when executing Go code.
	LastGoSP      uintptr // Last known Go stack pointer, used for detecting stack changes.
	GoBP          uintptr // Base pointer when executing Go code.
	SavedG        uintptr // Pointer to the G struct.

	// Execution state related fields.
	ReturnAddressAnchor       uintptr
	GlobalStubContext         uintptr    // Pointer to RegContext struct.
	GlobalExceptionInfo       uintptr    // On Windows: Pointer to EXCEPTION_POINTERS struct.
	GlobalExceptionInfoBuffer [2]uintptr // On Linux: Buffer to store siginfo_t* and ucontext_t* pointers.

	// Saved state for Call function (callee-saved registers).
	CallSavedBX  uintptr
	CallSavedR12 uintptr
	CallSavedR13 uintptr
	CallSavedR14 uintptr
	CallSavedR15 uintptr
}

// NewThreadContext creates a new ThreadContext for given thread ID and stack pointer.
func NewThreadContext(threadPtr, stackPtr uintptr) *ThreadContext {
	threadContext := &ThreadContext{
		ThreadPtr:     threadPtr,
		PlaystationSP: stackPtr,
	}
	ThreadContextLock.Lock()
	ThreadContextRepo[threadPtr] = threadContext
	ThreadContextLock.Unlock()
	ThreadContextPinner.Pin(threadContext)

	return threadContext
}

// GetCurrentThreadContext returns ThreadContext for the current thread.
func GetCurrentThreadContext() *ThreadContext

// SetThreadContext sets ThreadContext for the current thread.
func SetThreadContext(ctx *ThreadContext) {
	sys_struct.SetTlsSlot(GoTlsSlot, uintptr(unsafe.Pointer(ctx)))
}

// AllocTlsSlots allocates TLS slots for Go and guest contexts.
func AllocTlsSlots() {
	GoTlsSlot, GoTlsOffset = sys_struct.AllocTlsSlot()
	PlaystationTlsSlot, PlaystationTlsOffset = sys_struct.AllocTlsSlot()
	if GoTlsSlot >= 64 || PlaystationTlsSlot >= 64 {
		panic("tls slot is too high, this is not supported")
	}
}
