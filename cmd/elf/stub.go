package elf

import (
	"encoding/binary"
	"reflect"
	"slices"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// GetSymbolAddressFunc defines the signature for a function that retrieves a symbol's address.
type GetSymbolAddressFunc func(s *ElfSymbol) (uintptr, bool)

// GetDefiningModuleFunc defines the signature for a function that retrieves a symbol's defining module.
type GetDefiningModuleFunc func(s *ElfSymbol) *Elf

// GetSymbolAddress is a global function to retrieve a symbol's address.
var GetSymbolAddress GetSymbolAddressFunc

// GetDefiningModule is a global function to retrieve a symbol's defining module.
var GetDefiningModule GetDefiningModuleFunc

// RegisterStub registers a new stub for a Go function, creating an assembly trampoline.
// It associates the stub with a library and symbol name.
func RegisterStub(libraryName, symbolName string, goFn any) asm.StubInfo {
	goFunc := reflect.ValueOf(goFn)
	stub := asm.StubInfo{
		LibraryName: libraryName,
		SymbolName:  symbolName,
		Address:     CreateTrampoline(goFunc.Pointer()),
		FuncValue:   goFunc,
		FuncType:    goFunc.Type(),
	}
	hashIndex := GetSymbolHashIndex(libraryName, symbolName)
	asm.Stubs[hashIndex] = stub
	asm.StubsMap[goFunc.Pointer()] = hashIndex
	asm.StubsTrampolineMap[stub.Address] = hashIndex
	logger.Printf(
		"Registered %s assembly trampoline at %s to Go function at %s...\n",
		color.Blue.Sprintf("%s:%s", libraryName, symbolName),
		color.Yellow.Sprintf("0x%X", stub.Address),
		color.Yellow.Sprintf("0x%X", goFunc.Pointer()),
	)

	return stub
}

// RegisterAssemblyStub registers a new stub for an assembly function.
// It associates the stub with a library and symbol name.
func RegisterAssemblyStub(libraryName, symbolName string, functionAddress uintptr) asm.StubInfo {
	stub := asm.StubInfo{
		LibraryName: libraryName,
		SymbolName:  symbolName,
		Address:     functionAddress,
	}
	hashIndex := GetSymbolHashIndex(libraryName, symbolName)
	asm.Stubs[hashIndex] = stub
	asm.StubsMap[functionAddress] = hashIndex
	asm.StubsTrampolineMap[stub.Address] = hashIndex
	logger.Printf(
		"Registered %s as assembly function at %s...\n",
		color.Blue.Sprintf("%s:%s", libraryName, symbolName),
		color.Yellow.Sprintf("0x%X", stub.Address),
	)

	return stub
}

// RegisterVariableStub registers a new stub for a global variable.
// It allocates memory for the variable and associates it with a library and symbol name.
func RegisterVariableStub(libraryName, symbolName string, size uintptr) asm.StubInfo {
	addr := GlobalGoAllocator.Malloc(size)
	hashIndex := GetSymbolHashIndex(libraryName, symbolName)
	stub := asm.StubInfo{
		LibraryName: libraryName,
		SymbolName:  symbolName,
		Address:     addr,
	}
	asm.Stubs[hashIndex] = stub

	return stub
}

// CreateTrampoline generates an assembly trampoline that calls the specified Go function.
func CreateTrampoline(goFuncAddr uintptr) uintptr {
	// Allocate executable memory for the trampoline.
	trampolineSize := uintptr(22) // MOV to RAX (10), MOV to R11 (10), JMP RAX (2)
	trampolineAddr, _ := sys_struct.AllocExecutableMemory(trampolineSize)

	// MOV stubAsm, RAX
	trampoline := []byte{0x48, 0xB8}
	trampoline = binary.LittleEndian.AppendUint64(trampoline, uint64(asm.StubAddr))

	// MOV $<goFuncAddr>, R11
	trampoline = append(trampoline, 0x49, 0xBB)
	trampoline = binary.LittleEndian.AppendUint64(trampoline, uint64(goFuncAddr))

	// JMP RAX
	trampoline = append(trampoline, 0xFF, 0xE0)

	// Write the trampoline to memory.
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(trampolineAddr)), trampolineSize),
		trampoline,
	)

	return trampolineAddr
}

// NativeFunctionNames lists names of functions that should not be stubbed.
var NativeFunctionNames = []string{
	"pthread_once",
	"scePthreadOnce",
	"sceKernelSetCallRecord",
}

// CanStubFunctionName checks if a given function name is eligible for stubbing.
func CanStubFunctionName(funcName string) bool {
	return !slices.Contains(NativeFunctionNames, funcName)
}
