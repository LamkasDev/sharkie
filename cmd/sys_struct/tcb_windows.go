//go:build windows

package sys_struct

import (
	"sync"
)

var (
	PlaystationTlsSlot   uintptr
	PlaystationTlsOffset uintptr
	PlaystationTlsOnce   sync.Once
)

func AllocPlaystationTlsSlot() {
	slot, _, err := TlsAlloc.Call()
	if slot == 0 {
		panic(err)
	}
	PlaystationTlsSlot = slot
	PlaystationTlsOffset = 0x1480 + slot*8
}

// TODO: FilterTcbAccess
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/cpu_patches.cpp#L81C13-L81C28

// TODO: GenerateTcbAccess
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/cpu_patches.cpp#L92
