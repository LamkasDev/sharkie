package app_content

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/app_content"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001740
// __int64 __fastcall sceAppContentTemporaryDataMount2(unsigned int, __int64)
func libSceAppContent_sceAppContentTemporaryDataMount2(option AppContentTemporaryDataOption, mountPointPtr Cstring) uintptr {
	if mountPointPtr == nil {
		logger.Printf("%-132s %s failed due to invalid mount point pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAppContentTemporaryDataMount2"),
		)
		return 0x80D90002
	}
	mountPoint := "/temp0"
	CString(mountPointPtr, mountPoint)

	logger.Printf("%-132s %s returned temporary mount point %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAppContentTemporaryDataMount2"),
		color.Blue.Sprint(mountPoint),
	)
	return 0
}
