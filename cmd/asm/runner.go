package asm

import (
	_ "unsafe"
)

//go:linkname runtime_entersyscall runtime.entersyscall
func runtime_entersyscall()

//go:linkname runtime_exitsyscall runtime.exitsyscall
func runtime_exitsyscall()

var (
	ProcEntersyscall = runtime_entersyscall
	ProcExitsyscall  = runtime_exitsyscall
)

func Run(entry, stackPtr, argsPtr, arg2 uintptr)
func Call(entry, stackPtr, arg1, arg2 uintptr)
