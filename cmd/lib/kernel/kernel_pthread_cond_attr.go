package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// 0x0000000000013720
// __int64 scePthreadCondattrInit()
func libKernel_scePthreadCondattrInit(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_condattr_init(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000136C0
// __int64 scePthreadCondattrDestroy()
func libKernel_scePthreadCondattrDestroy(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_condattr_destroy(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
