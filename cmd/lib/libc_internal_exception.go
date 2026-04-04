package lib

// 0x00000000000CA900
// void __noreturn _cxa_throw(void *, struct type_info *lptinfo, void (*)(void *))
func libSceLibcInternal___cxa_throw(exceptionPtr uintptr) uintptr {
	return 0
}

// 0x0000000000054860
// void __noreturn std::_Xbad_alloc(void)
func libSceLibcInternal_std_Xbad_alloc() uintptr {
	return 0
}
