package emu

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
	. "github.com/LamkasDev/sharkie/cmd/structs/tcb"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

var (
	// ThreadRepo maps thread IDs to host threads.
	ThreadRepo = map[int32]*Thread{}

	// ThreadLock protects ThreadRepo, so multiple threads can look up threads safely.
	ThreadLock sync.RWMutex

	MainThreadId = NextThreadId
	NextThreadId = int32(1001)
)

type Thread struct {
	Id        int32
	Name      string
	Stack     *Stack
	Tcb       *Tcb
	KeyValues map[uint32]uintptr
	Lock      sync.Mutex

	IsMain   bool
	Exited   bool
	ExitCode uintptr
	JoinCond *sync.Cond

	SignalMask   ThreadSignalMask
	AffinityMask ThreadAffinityMask
}

func NewThread(name string, stackSize uint64) *Thread {
	thread := &Thread{
		Id:        NextThreadId,
		Stack:     NewStack(stackSize),
		KeyValues: map[uint32]uintptr{},
		Lock:      sync.Mutex{},
	}
	if thread.Id == MainThreadId {
		thread.IsMain = true
	}
	if name == "" {
		thread.Name = fmt.Sprintf("Thread-%d", thread.Id)
	} else {
		thread.Name = strings.ReplaceAll(name, "\n", "")
	}
	thread.Tcb = NewTcb(thread)
	thread.JoinCond = sync.NewCond(&thread.Lock)

	return thread
}

func CreateThread(name string, stackSize uint64) *Thread {
	thread := NewThread(name, stackSize)
	logger.Printf(
		"[%s] Stack of %s bytes allocated at %s (top %s).\n",
		color.Green.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", stackSize),
		color.Yellow.Sprintf("0x%X", thread.Stack.Address),
		color.Yellow.Sprintf("0x%X", thread.Stack.Top),
	)
	logger.Printf(
		"[%s] TCB allocated at %s (TLS at %s, %s bytes).\n",
		color.Green.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(thread.Tcb))),
		color.Yellow.Sprintf("0x%X", uint64(uintptr(unsafe.Pointer(thread.Tcb)))-linker.GlobalLinker.StaticTlsSize),
		color.Gray.Sprint(linker.GlobalLinker.StaticTlsSize),
	)

	ThreadLock.Lock()
	ThreadRepo[thread.Id] = thread
	NextThreadId++
	ThreadLock.Unlock()
	return thread
}

func GetCurrentThread() *Thread {
	threadContext := asm.GetCurrentThreadContext()
	return (*Thread)(unsafe.Pointer(threadContext.ThreadPtr))
}

func GetThreadForPtr(threadPtr uintptr) *Thread {
	for _, thread := range ThreadRepo {
		if thread.Tcb.Thread.Self == threadPtr {
			return thread
		}
	}

	return nil
}

// Setup sets the current thread's context and TLS.
func (t *Thread) Setup() {
	asm.SetThreadContext(asm.NewThreadContext(uintptr(unsafe.Pointer(t)), t.Stack.CurrentPointer))
	sys_struct.SetTlsSlot(asm.PlaystationTlsSlot, uintptr(unsafe.Pointer(t.Tcb)))
}

// CallUnsafe calls a function at specified address.
// The stack it returns on is no longer expandable as it might have split during guest execution.
// Use Call to avoid this behaviour.
func (t *Thread) CallUnsafe(funcAddr uintptr, arg uintptr) {
	// Call the assembly trampoline and call funcAddr function.
	asm.GuestEnter()
	asm.Call(funcAddr, t.Stack.CurrentPointer, arg, 0)
	asm.GuestLeave()
}

// Call sets up the current goroutine and calls a function at specified address.
// The stack it returns on is no longer expandable as it might have split during guest execution.
// This however doesn't matter as long as you use it within a fresh goroutine.
// It's aimed to be used asynchronously like 'go Call(...)' or for a more complete solution see CallAndWait.
func (t *Thread) Call(funcAddr uintptr, arg uintptr) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	t.Setup()
	t.CallUnsafe(funcAddr, arg)
}

// CallAndWait creates a new goroutine, calls a function at specified address and waits until it finishes.
// The stack it returns on is guaranteed to be safe even after guest execution.
func (t *Thread) CallAndWait(funcAddr uintptr, arg uintptr) {
	var wg sync.WaitGroup
	wg.Go(func() {
		t.Call(funcAddr, arg)
	})
	wg.Wait()
}

// Run creates a new goroutine, pushes arguments on stack and calls the program's entry point.
// It is non-blocking (we shouldn't return from there anyway).
func (t *Thread) Run(e *elf.Elf) {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		t.Setup()

		// Push program arguments to the stack.
		// int main(int argc, char* argv[])
		strAddr := t.Stack.PushString(fmt.Sprintf("%s\x00", e.Name))
		argsPtr := t.Stack.PushUint32(1)
		t.Stack.PushUint64(uint64(strAddr))
		t.Stack.PushUint64(0)

		// Call the assembly trampoline and jump into game code.
		entry := e.BaseAddress + uintptr(e.EntryAddress)
		logger.Printf(
			"Jumping to %s's entry point %s (relative=%s)...\n",
			color.Blue.Sprintf("%s", e.Name),
			color.Yellow.Sprintf("0x%X", entry),
			color.Yellow.Sprintf("0x%X", e.EntryAddress),
		)
		asm.GuestEnter()
		asm.Run(entry, t.Stack.CurrentPointer, argsPtr, 0)
		asm.GuestLeave()

		// This should not be reached.
		logger.Println("Returned from run - this should not happen.")
	}()
}

func (t *Thread) Exit(exitCode uintptr) {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	if t.Exited {
		return
	}

	// Mark thread as done.
	t.Exited = true
	t.ExitCode = exitCode
	t.Tcb.Thread.ReturnValue = exitCode

	// Process cleanup handlers.
	cleanupHandlerPtr := t.Tcb.Thread.CleanupStack
	if cleanupHandlerPtr != 0 {
		limit := 0
		for limit < 20 {
			entry := (*PthreadCleanupEntry)(unsafe.Pointer(cleanupHandlerPtr))
			module := GetModuleAtAddress(entry.Handler)
			if module == nil {
				logger.Printf("Thread %s failed finding cleanup handler at %s...\n",
					color.Blue.Sprint(t.Name),
					color.Yellow.Sprintf("0x%X", entry.Handler),
				)
				continue
			}
			logger.Printf("Thread %s skipped cleanup handler %s/%s with %s argument...\n",
				color.Blue.Sprint(t.Name),
				color.Blue.Sprint(module.Name),
				color.Yellow.Sprintf("0x%X", entry.Handler-module.BaseAddress),
				color.Yellow.Sprintf("0x%X", entry.Arg),
			)
			cleanupHandlerPtr = entry.Next
			limit++
		}
	} else {
		logger.Printf("Thread %s skipping empty cleanup handlers...\n",
			color.Blue.Sprint(t.Name),
		)
	}
	t.Tcb.Thread.CleanupStack = 0

	// Signal waiting threads.
	t.JoinCond.Broadcast()

	logger.Printf(
		"Thread %s exited with code %s.\n",
		color.Blue.Sprint(t.Name),
		color.Yellow.Sprintf("0x%X", t.ExitCode),
	)
}

// SafeReadUint64 safely reads a uint64 value from the stack.
func (t *Thread) SafeReadUint64(address uintptr) (uint64, bool) {
	if t.Stack != nil {
		if address >= t.Stack.Address && address+8 <= t.Stack.Address+uintptr(t.Stack.Size) {
			return *(*uint64)(unsafe.Pointer(address)), true
		}
	}

	return 0, false
}
