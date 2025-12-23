package elf

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

var FakeAddressStart = uint64(0xDEAD00000000)
var FakeAddress = FakeAddressStart
var FakeAddressMap = make(map[uint64]string)

type GetSymbolAddressFunc func(s *ElfSymbol) (uint64, bool)
type GetDefiningModuleFunc func(s *ElfSymbol) *Elf

var GetSymbolAddress GetSymbolAddressFunc
var GetDefiningModule GetDefiningModuleFunc

// RegisterStub registers a new stub specified by library and symbol name pointing to function f.
func RegisterStub(libraryName, symbolName string, f interface{}) asm.StubInfo {
	fn := reflect.ValueOf(f)
	stub := asm.StubInfo{
		LibraryName: libraryName,
		SymbolName:  symbolName,
		Address:     CreateTrampoline(fn.Pointer()),
		FuncValue:   fn,
		FuncType:    fn.Type(),
	}
	hashIndex := GetSymbolHashIndex(libraryName, symbolName)
	asm.Stubs[hashIndex] = stub
	asm.StubsMap[fn.Pointer()] = hashIndex
	asm.StubsTrampolineMap[stub.Address] = hashIndex
	fmt.Printf(
		"Registered %s assembly trampoline at %s to Go function at %s...\n",
		color.Blue.Sprintf("%s:%s", libraryName, symbolName),
		color.Yellow.Sprintf("0x%X", stub.Address),
		color.Yellow.Sprintf("0x%X", fn.Pointer()),
	)

	return stub
}

// RegisterVariableStub registers a new variable stub specified by library and symbol name of size.
func RegisterVariableStub(libraryName, symbolName string, size uintptr) asm.StubInfo {
	addr, _ := sys_struct.AllocReadWriteMemory(size)
	hashIndex := GetSymbolHashIndex(libraryName, symbolName)
	stub := asm.StubInfo{
		LibraryName: libraryName,
		SymbolName:  symbolName,
		Address:     addr,
	}
	asm.Stubs[hashIndex] = stub

	return stub
}

// CreateTrampoline generates a trampoline for a given Go function.
func CreateTrampoline(goFuncAddr uintptr) uintptr {
	// Allocate executable memory for the trampoline.
	trampolineSize := uintptr(22) // MOV to RAX (10), MOV to R11 (10), JMP RAX (2)
	trampolineAddr, _ := sys_struct.AllocExecututableMemory(trampolineSize)

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
