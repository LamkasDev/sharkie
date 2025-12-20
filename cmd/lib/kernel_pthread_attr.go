package lib

// 0x00000000000134A0
// __int64 __fastcall scePthreadAttrInit(__int64 *)
func libKernel_scePthreadAttrInit(attrHandlePtr uintptr) uintptr {
	err := libKernel_pthread_attr_init(attrHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}
