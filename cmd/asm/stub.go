package asm

import (
	"reflect"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
)

const RegContextSize = 384

type StubInfo struct {
	LibraryName string
	SymbolName  string
	Address     uintptr
	FuncValue   reflect.Value
	FuncType    reflect.Type
}

var Stubs = make(map[uint64]StubInfo)
var StubsMap = make(map[uintptr]uint64)
var StubsTrampolineMap = make(map[uintptr]uint64)

func InitStubAddr()

// stubGo is a Go function that acts as a trampoline to call the target function.
// It is called from stubAsm.
func stubGo() {
	// GuestLeave()
	// defer GuestEnter()
	// CheckAndRunGC()

	// Extract arguments and function pointer from RegSaveArea.
	threadContext := GetCurrentThreadContext()
	ctx := (*RegContext)(unsafe.Pointer(threadContext.GlobalStubContext))
	fnPtr := ctx.R11

	if threadContext.LastGoSP != threadContext.GoSP {
		logger.Printf("Stack changed from 0x%X to 0x%X.\n", threadContext.LastGoSP, threadContext.GoSP)
	}

	// Look up the stub info.
	stubName, ok := StubsMap[fnPtr]
	if !ok {
		panic("stub not found")
	}
	stubInfo := Stubs[stubName]

	// Prepare arguments.
	arguments := make([]reflect.Value, stubInfo.FuncType.NumIn())
	for i := 0; i < len(arguments); i++ {
		argType := stubInfo.FuncType.In(i)
		argVal := reflect.New(argType).Elem()
		switch i {
		case 0:
			argVal.SetUint(uint64(ctx.DI))
			break
		case 1:
			argVal.SetUint(uint64(ctx.SI))
			break
		case 2:
			argVal.SetUint(uint64(ctx.DX))
			break
		case 3:
			argVal.SetUint(uint64(ctx.CX))
			break
		case 4:
			argVal.SetUint(uint64(ctx.R8))
			break
		case 5:
			argVal.SetUint(uint64(ctx.R9))
			break
		default:
			// RegContextSize + stubAsm return address + original return address + offset.
			stackOffset := RegContextSize + 8 + uintptr((i-6)*8)
			addr := (*uint64)(unsafe.Pointer(uintptr(unsafe.Pointer(ctx)) + stackOffset))
			argVal.SetUint(*addr)
		}
		arguments[i] = argVal
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
