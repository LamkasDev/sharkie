//go:build linux

package sys_struct

/*
	#include <pthread.h>
*/
import "C"
import (
	"fmt"
	"syscall"
	"unsafe"
)

func AllocTlsSlot() (uintptr, uintptr) {
	var slot C.pthread_key_t
	err := C.pthread_key_create(&slot, nil)
	if err != 0 {
		panic("pthread_key_create failed")
	}
	offset := GetTlsSlotOffset(slot)

	return uintptr(slot), offset
}

func GetTlsSlotOffset(slot C.pthread_key_t) (offset uintptr) {
	// Set magic to slot.
	magic := uintptr(0xDEADBEEFCAFEBABE)
	C.pthread_setspecific(slot, unsafe.Pointer(magic))

	// Get FS segment base address (ARCH_GET_FS = 0x1003).
	var base uintptr
	_, _, err := syscall.Syscall(syscall.SYS_ARCH_PRCTL, 0x1003, uintptr(unsafe.Pointer(&base)), 0)
	if err != 0 {
		panic(fmt.Sprintf("arch_prctl failed: %v", err))
	}

	// Scan memory to find our magic value.
	found := false
	for ; offset < 4096; offset += 8 {
		// Read memory at FS + offset
		value := *(*uintptr)(unsafe.Pointer(base + offset))
		if value == magic {
			found = true
			break
		}
	}
	if !found {
		panic("failed to locate TLS slot offset in FS segment")
	}

	// Clear magic.
	C.pthread_setspecific(slot, nil)

	return offset
}

func SetTlsSlot(slot uintptr, value uintptr) {
	status := C.pthread_setspecific(C.pthread_key_t(slot), unsafe.Pointer(value))
	if status != 0 {
		panic("pthread_setspecific failed")
	}
}
