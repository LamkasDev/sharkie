package save_data

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000276F0
// __int64 __fastcall sceSaveDataInitialize(__int64)
func libSceSaveData_sceSaveDataInitialize(param uintptr) uintptr {
	logger.Printf("sceSaveDataInitialize called (param=%x)\n", param)
	return 0
}

// 0x0000000000028F00
// __int64 __fastcall sceSaveDataDirNameSearch(int, int)
func libSceSaveData_sceSaveDataDirNameSearch(param uintptr) uintptr {
	logger.Printf("sceSaveDataDirNameSearch called (param=%x)\n", param)
	return 0
}

// 0x0000000000027970
// __int64 __fastcall sceSaveDataMount(__int64, int)
func libSceSaveData_sceSaveDataMount(mountPtr, resultPtr uintptr) uintptr {
	if mountPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mount pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0000
	}

	mount := (*OrbisSaveDataMount)(unsafe.Pointer(mountPtr))
	if mount.UserId < 0 {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0011
	}
	if mount.DirName == nil {
		logger.Printf("%-132s %s failed due to invalid dir name.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0000
	}
	create := mount.MountMode&OrbisSaveDataMountModeCREATE != 0
	createIfNotExists := mount.MountMode&OrbisSaveDataMountModeCREATE2 != 0
	copyIcon := mount.MountMode&OrbisSaveDataMountModeCOPY_ICON != 0

	// Get next available mount slot.
	mountSlot, err := GlobalSaveDataManager.GetAvailableMountSlot(GoString(mount.DirName))
	if err != nil {
		logger.Printf("%-132s %s failed due to slot error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
			err.Error(),
		)
		return 0x809F0003
	}

	// Create a new save instance.
	saveInstance := NewSaveInstance(mount.UserId, GoString(mount.TitleId), GoString(mount.DirName), uint64(mount.Blocks))
	exists := saveInstance.ExistsOnHost()
	if !create && !createIfNotExists && !exists {
		logger.Printf("%-132s %s failed due to no existing save (no create flag).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0008
	}
	if create && exists {
		logger.Printf("%-132s %s failed due to existing save (with create flag).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0007
	}

	// Mount save instance at slot (will create if it's new).
	mountPoint, err, created := saveInstance.Mount(mountSlot, copyIcon)
	if err != nil {
		logger.Printf("%-132s %s failed due to mount error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
			err.Error(),
		)
		return 0x809F000B
	}

	// Set mount result.
	mountResult := (*OrbisSaveDataMountResult)(unsafe.Pointer(resultPtr))
	CString(mountResult.MountPoint, mountPoint)
	if created {
		mountResult.MountStatus = OrbisSaveDataMountStatusCREATED
	}
	GlobalSaveDataManager.SetSaveInstance(mountSlot, saveInstance)

	logger.Printf("%-132s %s mounted save at %s (userId=%s, titleId=%s, dirName=%s, blocks=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataMount"),
		color.Green.Sprint(mountPoint),
		color.Yellow.Sprintf("0x%X", mount.UserId),
		color.Yellow.Sprintf("0x%X", mount.TitleId),
		color.Green.Sprint(GoString(mount.DirName)),
		color.Yellow.Sprint(mount.Blocks),
	)
	return 0
}

// 0x0000000000028120
// __int64 __fastcall sceSaveDataUmount(int)
func libSceSaveData_sceSaveDataUmount(mountPointPtr uintptr) uintptr {
	logger.Printf("sceSaveDataUmount called (mountPointPtr=%x)\n", mountPointPtr)
	return 0
}

// 0x0000000000028C20
// __int64 __fastcall sceSaveDataGetMountInfo(int, int)
func libSceSaveData_sceSaveDataGetMountInfo(infoPtr uintptr) uintptr {
	logger.Printf("sceSaveDataGetMountInfo called (infoPtr=%x)\n", infoPtr)
	return 0
}

// 0x0000000000029200
// __int64 __fastcall sceSaveDataGetParam(int, int, int, __int64, __int64)
func libSceSaveData_sceSaveDataGetParam(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataGetParam called (paramPtr=%x)\n", paramPtr)
	return 0
}

// 0x0000000000029140
// __int64 __fastcall sceSaveDataSetParam(int, int, int, __int64)
func libSceSaveData_sceSaveDataSetParam(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataSetParam called (paramPtr=%x)\n", paramPtr)
	return 0
}

// 0x00000000000292C0
// __int64 __fastcall sceSaveDataSaveIcon(int, int)
func libSceSaveData_sceSaveDataSaveIcon(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataSaveIcon called (paramPtr=%x)\n", paramPtr)
	return 0
}

// 0x0000000000028CD0
// __int64 __fastcall sceSaveDataDelete(int)
func libSceSaveData_sceSaveDataDelete(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataDelete called (paramPtr=%x)\n", paramPtr)
	return 0
}
