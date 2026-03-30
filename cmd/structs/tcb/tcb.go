package tcb

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs/pthread"
)

const TcbAlignment = 64

// DtvEntry represent an entry in a dynamic thread vector.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L22
type DtvEntry struct {
	Counter uintptr
	Pointer uintptr
}

const DtvEntrySize = unsafe.Sizeof(DtvEntry{})

// Tcb represent the thread control block used by a thread.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L27
type Tcb struct {
	Self   *Tcb
	Dtv    *DtvEntry
	Thread *pthread.Pthread
	Fiber  uintptr
}

const TcbSize = unsafe.Sizeof(Tcb{})

// TlsIndex represents a request for a TLS base address.
type TlsIndex struct {
	ModuleId uint64
	Offset   uintptr
}

const TlsIndexSize = unsafe.Sizeof(TlsIndex{})
