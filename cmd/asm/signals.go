package asm

type ExceptionHandlerFunc func() uintptr

var ExceptionHandler ExceptionHandlerFunc

func InitSignalsAddr()
func exceptionHandlerGo() uintptr {
	return ExceptionHandler()
}
