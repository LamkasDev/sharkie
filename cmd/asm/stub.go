package asm

import (
	"reflect"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
)

// StubInfo holds information about a stubbed function.
type StubInfo struct {
	LibraryName string
	SymbolName  string
	Address     uintptr
	FuncValue   reflect.Value
	FuncType    reflect.Type
}

// Stubs stores information about stubbed functions, indexed by a calculated hash.
var Stubs = make(map[uint64]StubInfo)

// StubsMap maps original function addresses to their corresponding hash indexes in Stubs.
var StubsMap = make(map[uintptr]uint64)

// StubsTrampolineMap maps addresses of the stub trampolines to their corresponding hash indexes in Stubs.
var StubsTrampolineMap = make(map[uintptr]uint64)

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
	stubName, ok := StubsMap[fnPtr]
	if !ok {
		panic("stub not found")
	}
	stubInfo := Stubs[stubName]

	// Prepare arguments for the Go function call.
	arguments := make([]reflect.Value, stubInfo.FuncType.NumIn())
	for argIndex := range arguments {
		argType := stubInfo.FuncType.In(argIndex)
		argValue := reflect.New(argType).Elem()

		// Map guest registers to Go function arguments.
		switch argIndex {
		case 0:
			argValue.SetUint(uint64(ctx.DI))
		case 1:
			argValue.SetUint(uint64(ctx.SI))
		case 2:
			argValue.SetUint(uint64(ctx.DX))
		case 3:
			argValue.SetUint(uint64(ctx.CX))
		case 4:
			argValue.SetUint(uint64(ctx.R8))
		case 5:
			argValue.SetUint(uint64(ctx.R9))
		default:
			// RegContextSize + stubAsm return address + original return address + offset.
			stackOffset := RegContextSize + 8 + uintptr((argIndex-6)*8)
			addr := (*uint64)(unsafe.Add(unsafe.Pointer(ctx), stackOffset))
			argValue.SetUint(*addr)
		}
		arguments[argIndex] = argValue
	}

	// Call the function.
	// fmt.Printf("[%d] %s:%s\n", threadContext.ThreadId, stubInfo.LibraryName, stubInfo.SymbolName)
	results := stubInfo.FuncValue.Call(arguments)

	// Return the result.
	if len(results) > 0 {
		ctx.AX = uintptr(results[0].Uint())
	}
	threadContext.LastGoSP = threadContext.GoSP
}
