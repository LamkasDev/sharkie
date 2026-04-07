package emu

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

func RegisterFiosStubs() {
	fiosModule := GlobalModuleManager.ModulesMap["libSceFios2.sprx"]
	originalSymbol := fiosModule.SymbolTable.SymbolsMap[elf.GetSymbolHashIndex("libSceFios2", "sceFiosInitialize")]
	originalAddress := fiosModule.BaseAddress + originalSymbol.Address

	elf.RegisterAssemblyStub("libSceFios2", "sceFiosInitialize", CreateFiosTrampoline(originalAddress))
}

// CreateFiosTrampoline creates an assembly trampoline to fix FIOS initializition.
// When a game tries to initialize using older 152-byte parameters, the library denies it.
// We extend the original 152-bytes to 176-bytes and update the header.
func CreateFiosTrampoline(originalAddress uintptr) uintptr {
	var trampoline []byte

	// SUB RSP, 0xE8         ; Allocate 232 bytes (for alignment)
	trampoline = append(trampoline, 0x48, 0x81, 0xEC, 0xE8, 0x00, 0x00, 0x00)
	// MOV RSI, RDI          ; Source = OldParams (Arg0)
	trampoline = append(trampoline, 0x48, 0x89, 0xFE)
	// LEA RDI, [RSP+0x20]   ; Dest = Stack + 32
	trampoline = append(trampoline, 0x48, 0x8D, 0x7C, 0x24, 0x20)
	// MOV RCX, 19           ; Copy 19 Qwords (152 bytes)
	trampoline = append(trampoline, 0x48, 0xC7, 0xC1, 0x13, 0x00, 0x00, 0x00)
	// REP MOVSQ             ; Execute copy
	trampoline = append(trampoline, 0xF3, 0x48, 0xA5)
	// XOR RAX, RAX          ; Zero RAX
	trampoline = append(trampoline, 0x48, 0x31, 0xC0)
	// MOV [RDI], RAX        ; Zero next 8 bytes (152-160)
	trampoline = append(trampoline, 0x48, 0x89, 0x07)
	// MOV [RDI+8], RAX      ; Zero next 8 bytes (160-168)
	trampoline = append(trampoline, 0x48, 0x89, 0x47, 0x08)
	// MOV [RDI+16], RAX     ; Zero next 8 bytes (168-176)
	trampoline = append(trampoline, 0x48, 0x89, 0x47, 0x10)

	// LEA RDI, [RSP+0x20]  ; Point RDI back to start of NewStruct
	trampoline = append(trampoline, 0x48, 0x8D, 0x7C, 0x24, 0x20)
	// MOV EAX, [RDI]       ; Read header
	trampoline = append(trampoline, 0x8B, 0x07)
	// AND EAX, 0xFFFF0001   ; Clear size bits 1-15 only
	trampoline = append(trampoline, 0x25, 0x01, 0x00, 0xFF, 0xFF)
	// OR EAX, 352          ; Set size 176 (176 << 1 = 352 = 0x160)
	trampoline = append(trampoline, 0x0D, 0x60, 0x01, 0x00, 0x00)
	// MOV [RDI], EAX       ; Write header
	trampoline = append(trampoline, 0x89, 0x07)

	// MOV RAX, <ADDR>      ; Load original function address
	trampoline = append(trampoline, 0x48, 0xB8)
	trampoline = binary.LittleEndian.AppendUint64(trampoline, uint64(originalAddress))
	// CALL RAX             ; Call original function
	trampoline = append(trampoline, 0xFF, 0xD0)
	// ADD RSP, 0xE8        ; Restore stack
	trampoline = append(trampoline, 0x48, 0x81, 0xC4, 0xE8, 0x00, 0x00, 0x00)
	// RET                  ; Return to game
	trampoline = append(trampoline, 0xC3)

	trampolineAddr, err := sys_struct.AllocExecutableMemory(uint64(len(trampoline)))
	if err != nil {
		panic(err)
	}

	copy(unsafe.Slice((*byte)(unsafe.Pointer(trampolineAddr)), len(trampoline)), trampoline)

	return trampolineAddr
}
