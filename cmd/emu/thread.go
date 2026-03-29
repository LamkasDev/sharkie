package emu

import (
	"errors"
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
	Id    int32
	Name  string
	Stack *Stack
	Tcb   *Tcb
	Lock  sync.Mutex

	IsMain   bool
	Exited   bool
	ExitCode uintptr
	JoinCond *sync.Cond

	SignalMask   ThreadSignalMask
	AffinityMask ThreadAffinityMask
}

func NewThread(namePtr, stackSize uintptr) *Thread {
	thread := &Thread{
		Id:    NextThreadId,
		Stack: NewStack(stackSize),
		Lock:  sync.Mutex{},
	}
	if namePtr == 0 {
		if thread.Id == MainThreadId {
			thread.Name = "MainThread"
			thread.IsMain = true
		} else {
			thread.Name = fmt.Sprintf("Thread-%d", thread.Id)
		}
	} else {
		thread.Name = strings.ReplaceAll(ReadCString(namePtr), "\n", "")
	}
	thread.Tcb = NewTcb(thread)
	thread.JoinCond = sync.NewCond(&thread.Lock)

	return thread
}

func CreateThread(namePtr, stackSize uintptr) *Thread {
	thread := NewThread(namePtr, stackSize)
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
	ThreadLock.RLock()
	defer ThreadLock.RUnlock()

	threadContext := asm.GetCurrentThreadContext()
	if threadContext == nil {
		panic(errors.New("unknown thread context"))
	}
	thread := ThreadRepo[int32(threadContext.ThreadId)]
	if thread == nil {
		panic(errors.New("unknown thread"))
	}

	return thread
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
	asm.SetThreadContext(asm.NewThreadContext(t.Id, t.Stack.CurrentPointer))
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
// It's aimed to be used asynchronously like 'go Call(...)' or for a more complete solution see CallSync.
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

// Run pushes arguments on stack and calls the program's entry point.
func (t *Thread) Run(e *elf.Elf) {
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
}

// SafeReadUint64 safely reads a uint64 value from the stack.
func (t *Thread) SafeReadUint64(address uintptr) (uint64, bool) {
	if t.Stack != nil {
		if address >= t.Stack.Address && address+8 <= t.Stack.Address+t.Stack.Size {
			return *(*uint64)(unsafe.Pointer(address)), true
		}
	}

	return 0, false
}
