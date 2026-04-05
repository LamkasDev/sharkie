package structs

import "C"
import "unsafe"

// Cstring represents a C-style null-terminated string (const char*).
type Cstring unsafe.Pointer

func CString(ptr Cstring, s string) {
	if ptr == nil {
		return
	}
	dest := unsafe.Slice((*byte)(ptr), len(s)+1)
	copy(dest, s)
	dest[len(s)] = 0
}

func GoString(s Cstring) string {
	if s == nil {
		return ""
	}
	return C.GoString((*C.char)(s))
}
