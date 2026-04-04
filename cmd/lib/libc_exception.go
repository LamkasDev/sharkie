package lib

// 0x00000000000B29B0
// void __noreturn _cxa_throw(void *, struct type_info *lptinfo, void (*)(void *))
func libLibc___cxa_throw(exceptionPtr uintptr) uintptr {
	return libSceLibcInternal___cxa_throw(exceptionPtr)
}

// 0x0000000000054860
// void __noreturn std::_Xbad_alloc(void)
func libLibc_std_Xbad_alloc() uintptr {
	return libSceLibcInternal_std_Xbad_alloc()
}
