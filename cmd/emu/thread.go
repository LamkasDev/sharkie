package emu

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
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

	// Clear a 128-byte red zone and align to 16-bytes.
	// https://wiki.osdev.org/System_V_ABI
	stackPtr := thread.Stack.ArgumentsAddress - 128
	stackPtr &^= 15
	logger.Printf(
		"[%s] Stack of %s bytes allocated at %s (top %s).\n",
		color.Green.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", stackSize),
		color.Yellow.Sprintf("0x%X", thread.Stack.Address),
		color.Yellow.Sprintf("0x%X", stackPtr),
	)
	thread.Stack.CurrentPointer = stackPtr

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

func (t *Thread) Setup() {
	asm.SetThreadContext(asm.NewThreadContext(t.Id, t.Stack.CurrentPointer))

	// Allocate and set up the TCB.
	tcbAddr := uintptr(unsafe.Pointer(t.Tcb))
	ret, _, err := sys_struct.TlsSetValue.Call(sys_struct.PlaystationTlsSlot, tcbAddr)
	if ret == 0 {
		panic(err)
	}
	logger.Printf(
		"[%s] TCB allocated at %s (TLS at %s, %s bytes).\n",
		color.Green.Sprint(t.Name),
		color.Yellow.Sprintf("0x%X", tcbAddr),
		color.Yellow.Sprintf("0x%X", uint64(tcbAddr)-linker.GlobalLinker.StaticTlsSize),
		color.Gray.Sprint(linker.GlobalLinker.StaticTlsSize),
	)
}

// Call calls function at specified address.
func (t *Thread) Call(funcAddr uintptr, arg uintptr) {
	stackPtr := t.Stack.CurrentPointer
	stackPtr &^= 15

	// Call the assembly trampoline and call funcAddr function.
	// asm.GuestEnter()
	asm.Call(funcAddr, stackPtr, arg, 0)
	// asm.GuestLeave()
}

// Run pushes arguments on stack and calls the program's entry point.
func (t *Thread) Run(e *elf.Elf) {
	// Push program arguments to the stack.
	// int main(int argc, char* argv[])
	strAddr := t.Stack.PushString(fmt.Sprintf("%s\x00", e.Name))
	argsPtr := t.Stack.PushUint32(1)
	t.Stack.PushUint64(uint64(strAddr))
	t.Stack.PushUint64(0)

	// Clear a 128-byte red zone and align to 16-bytes.
	// https://wiki.osdev.org/System_V_ABI
	stackPtr := t.Stack.ArgumentsAddress - 128
	stackPtr &^= 15

	// Call the assembly trampoline and jump into game code.
	entry := e.BaseAddress + uintptr(e.EntryAddress)
	logger.Printf(
		"Jumping to entry point %s...\n",
		color.Yellow.Sprintf("0x%X", entry),
	)
	// asm.GuestEnter()
	asm.Run(entry, stackPtr, argsPtr, 0)
	// asm.GuestLeave()

	// This should not be reached.
	logger.Println("Returned from run - this should not happen.")
}

// SafeReadUint64 safely reads a uint64 value from the stack.
func (t *Thread) SafeReadUint64(address uintptr) (uint64, bool) {
	if t.Stack != nil {
		if address >= t.Stack.Address && address+8 <= t.Stack.Address+uintptr(len(t.Stack.Contents)) {
			return *(*uint64)(unsafe.Pointer(address)), true
		}
	}

	return 0, false
}
