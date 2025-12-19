package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/mem"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// Prepare creates a new stack and TCB before any code runs.
func (m *ModuleManager) Prepare(l *linker.Linker) {
	m.Stack = mem.NewStack(uintptr(mem.StackDefaultSize))

	// Clear a 128-byte red zone and align to 16-bytes.
	// https://wiki.osdev.org/System_V_ABI
	stackPtr := m.Stack.ArgumentsAddress - 128
	stackPtr &^= 15
	fmt.Printf(
		"Stack allocated at %s (top %s).\n",
		color.Yellow.Sprintf("0x%X", m.Stack.Address),
		color.Yellow.Sprintf("0x%X", stackPtr),
	)
	m.Stack.CurrentPointer = stackPtr

	// Allocate and set up the TCB
	m.Tcb = NewTCB(l)
	tcbAddr := uintptr(unsafe.Pointer(m.Tcb))
	sys_struct.TlsSetValue.Call(sys_struct.TlsSlot, tcbAddr)
	fmt.Printf(
		"TCB allocated at %s (TLS at %s, %s bytes).\n",
		color.Yellow.Sprintf("0x%X", tcbAddr),
		color.Yellow.Sprintf("0x%X", uint64(tcbAddr)-l.StaticTlsSize),
		color.Gray.Sprint(l.StaticTlsSize),
	)
}

// Call calls function at specified address.
func (m *ModuleManager) Call(funcAddr uintptr) {
	stackPtr := m.Stack.CurrentPointer
	stackPtr &^= 15

	// Call the assembly trampoline and call funcAddr function.
	asm.Call(funcAddr, stackPtr, 0, 0)
}

// Run creates a new stack and calls the program's entry point.
func (m *ModuleManager) Run(e *elf.Elf) {
	// Push program arguments to the stack.
	// int main(int argc, char* argv[])
	m.Stack.PushUint32(1)
	m.Stack.PushUint64(uint64(m.Stack.ArgumentsAddress + 16))
	m.Stack.PushString(fmt.Sprintf("%s\x00", e.Name))

	// Clear a 128-byte red zone and align to 16-bytes.
	// https://wiki.osdev.org/System_V_ABI
	stackPtr := m.Stack.ArgumentsAddress - 128
	stackPtr &^= 15

	// Call the assembly trampoline and jump into game code.
	entry := e.BaseAddress + uintptr(e.EntryAddress)
	fmt.Printf(
		"Jumping to entry point %s...\n",
		color.Yellow.Sprintf("0x%X", entry),
	)
	asm.Run(entry, stackPtr, 0, 0)

	// This should not be reached.
	fmt.Println("Returned from run - this should not happen.")
}
