//go:build windows

package sys_struct

func AllocTlsSlot() (uintptr, uintptr) {
	slot, _, err := sys_struct.TlsAlloc.Call()
	if slot == 0 {
		panic(err)
	}

	return slot, 0x1480 + slot*8
}

func SetTlsSlot(slot uintptr, value uintptr) {
	status, _, err := sys_struct.TlsSetValue.Call(slot, value)
	if status == 0 {
		panic(err)
	}
}
