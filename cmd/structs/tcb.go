package structs

// DtvEntry represent an entry in a dynamic thread vector.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L22
type DtvEntry struct {
	Counter uintptr
	Pointer uintptr
}

// Tcb represent the thread control block used by a thread.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L27
type Tcb struct {
	Self   *Tcb
	Dtv    *DtvEntry
	Thread *Pthread
	Fiber  uintptr
}

// TlsIndex represents a request for a TLS base address.
type TlsIndex struct {
	ModuleId uint64
	Offset   uintptr
}

var (
	// TlsBaseRepo maps module indexes to host TLS base addresses.
	TlsBaseRepo = map[uint64]uintptr{}
)
