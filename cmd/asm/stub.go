package asm

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
)

// StubDispatcher defines the signature for stubbed function call.
type StubDispatcher func(ctx *RegContext) uintptr

// StubInfo holds information about a stubbed function.
type StubInfo struct {
	LibraryName string
	SymbolName  string
	Address     uintptr
	Dispatcher  StubDispatcher
}

// Stubs stores information about stubbed functions, indexed by a calculated hash.
var Stubs = make(map[uint64]*StubInfo)

// StubsMap maps original function addresses to their corresponding stubs in Stubs.
var StubsMap = make(map[uintptr]*StubInfo)

// StubsTrampolineMap maps addresses of the stub trampolines to their corresponding stubs in Stubs.
var StubsTrampolineMap = make(map[uintptr]*StubInfo)

// InitStubAddr initializes the address of the stub assembly function.
func InitStubAddr()

// stubGo is a trampoline to call the target Go function from stubAsm.
func stubGo() {
	GuestLeave()
	defer GuestEnter()
	CheckAndRunGC()

	// Extract arguments and function pointer from RegSaveArea.
	threadContext := GetCurrentThreadContext()
	ctx := (*RegContext)(unsafe.Pointer(threadContext.GlobalStubContext))
	fnPtr := ctx.R11

	if threadContext.LastGoSP != threadContext.GoSP {
		logger.Printf("Stack changed from 0x%X to 0x%X.\n", threadContext.LastGoSP, threadContext.GoSP)
	}

	// Look up stub info using the function pointer.
	stubInfo := StubsMap[fnPtr]

	// Call the function.
	// fmt.Printf("[%d] %s:%s\n", threadContext.ThreadId, stubInfo.LibraryName, stubInfo.SymbolName)
	ctx.AX = stubInfo.Dispatcher(ctx)

	threadContext.LastGoSP = threadContext.GoSP
}
