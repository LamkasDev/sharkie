package disc_map

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// 0x0000000000001140
// __int64 __fastcall sceDiscMapGetPackageSize(__int64, unsigned __int64 *, _QWORD *)
func libSceSceDiscMap_sceDiscMapGetPackageSize() uintptr {
	return 0x81100004
}

// 0x00000000000012F0
// __int64 __fastcall sceDiscMapIsRequestOnHDD(__int64, __int64, __int64, _DWORD *)
func libSceSceDiscMap_sceDiscMapIsRequestOnHDD() uintptr {
	return 0x81100004
}

// 0x0000000000000FF0
// __int64 __fastcall fJgP_wqifno_C_A(__int64, __int64, __int64, _QWORD *, _QWORD *, _QWORD *)
func libSceSceDiscMap_fjg(path Cstring, offset, numBytes int64, flags, ret1, ret2 *int32) uintptr {
	*flags = 0
	*ret1 = 0
	*ret2 = 0
	return 0
}
