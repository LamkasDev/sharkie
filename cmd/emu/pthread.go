package emu

type Pthread struct {
	Magic               uint32
	ThreadId            int32
	Flags               uint32
	_                   uint32 // Padding.
	ReturnValue         uintptr
	Error               int32
	_                   int32 // More padding yippee!
	CleanupHandlerStack uintptr
	Name                [32]byte
}

const PthreadMagic = 0xDEADBEEF
