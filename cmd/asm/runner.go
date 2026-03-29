// Package asm handles low-level emulation setup.
package asm

// Run switches to the game's stack and jumps to specified entry point.
// It does not return.
func Run(entry, stackPtr, argsPtr, arg2 uintptr)

// Call switches to the game's stack, calls a function at specified entry point and returns.
// We can't expand the caller's stack afterward or there will be trouble (split-stack overflow).
func Call(entry, stackPtr, arg1, arg2 uintptr)
