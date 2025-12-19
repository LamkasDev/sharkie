package asm

type ExceptionHandlerFunc func() uintptr

var ExceptionHandler ExceptionHandlerFunc

var (
	// ExceptionHandlerAddr holds address of assembly exception handler function.
	ExceptionHandlerAddr uintptr

	// GlobalExceptionInfo holds address of the exception info struct.
	GlobalExceptionInfo uintptr
)

func InitSignalsAddr()
func exceptionHandlerGo() uintptr {
	return ExceptionHandler()
}
