package asm

import (
	"unsafe"
)

type StubInfo struct {
	LibraryName string
	SymbolName  string
	Address     uintptr
	NumArgs     int
}

var (
	// StubAddr holds address of the assembly stub.
	StubAddr uintptr

	// GlobalStubContext holds address of the RegContext struct.
	GlobalStubContext uintptr
)

var Stubs = make(map[uint64]StubInfo)
var StubsMap = make(map[uintptr]uint64)
var StubsTrampolineMap = make(map[uintptr]uint64)

func InitStubAddr()

// stubGo is a Go function that acts as a trampoline to call the target function.
// It is called from stubAsm.
func stubGo() {
	// Extract arguments and function pointer from RegSaveArea.
	ctx := (*RegContext)(unsafe.Pointer(GlobalStubContext))
	fnPtr := ctx.R11
	arg1 := ctx.DI
	arg2 := ctx.SI
	arg3 := ctx.DX
	arg4 := ctx.CX
	arg5 := ctx.R8
	arg6 := ctx.R9
	arg7 := ctx.R10

	// Look up the stub info.
	stubName, ok := StubsMap[fnPtr]
	if !ok {
		panic("stub not found")
	}
	stubInfo := Stubs[stubName]

	// Call the function with the correct number of arguments.
	var result uintptr
	ptrToFnPtr := &fnPtr
	switch stubInfo.NumArgs {
	case 0:
		targetFunc := *(*func() uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc()
	case 1:
		targetFunc := *(*func(uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1)
	case 2:
		targetFunc := *(*func(uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2)
	case 3:
		targetFunc := *(*func(uintptr, uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2, arg3)
	case 4:
		targetFunc := *(*func(uintptr, uintptr, uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2, arg3, arg4)
	case 5:
		targetFunc := *(*func(uintptr, uintptr, uintptr, uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2, arg3, arg4, arg5)
	case 6:
		targetFunc := *(*func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2, arg3, arg4, arg5, arg6)
	case 7:
		targetFunc := *(*func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr)(unsafe.Pointer(&ptrToFnPtr))
		result = targetFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	default:
		panic("unsupported number of arguments")
	}

	// Return the result.
	ctx.AX = result
}
