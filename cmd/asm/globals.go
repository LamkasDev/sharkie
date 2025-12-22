package asm

var (
	StubAddr             uintptr
	ExceptionHandlerAddr uintptr
	GlobalExceptionInfo  uintptr
	GlobalStubContext    uintptr

	WindowsStackSP     uintptr
	PlaystationStackSP uintptr
	GoStackSP          uintptr
	GoStackBP          uintptr
	SavedG             uintptr

	ReturnAddressAnchor uintptr
	CallReturnAddress   uintptr
)
