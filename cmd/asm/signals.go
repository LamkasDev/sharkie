package asm

// ExceptionHandlerFunc defines signature for a function that handles exceptions.
type ExceptionHandlerFunc func() uintptr

// ExceptionHandler holds the currently registered exception handler.
var ExceptionHandler ExceptionHandlerFunc

// InitSignalsAddr initializes address of the signal handling assembly function.
func InitSignalsAddr()

// exceptionHandlerGo is a Go function that acts as a trampoline for the exception handler.
// It is called from assembly code when an exception occurs.
func exceptionHandlerGo() uintptr {
	return ExceptionHandler()
}
