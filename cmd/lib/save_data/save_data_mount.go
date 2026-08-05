package save_data

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000027970
// __int64 __fastcall sceSaveDataMount(__int64, int)
func libSceSaveData_sceSaveDataMount(mount *SaveDataMount, mountResult *SaveDataMountResult) uintptr {
	if mount == nil {
		logger.Printf("%-132s %s failed due to invalid mount pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
		)
		return 0x809F0000
	}
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

	return saveDataMount(mount, mountResult)
}

// 0x0000000000027C50
// __int64 __fastcall sceSaveDataMount2(__int64, int, __m128 _XMM0)
func libSceSaveData_sceSaveDataMount2(mount2 *SaveDataMount2, mountResult *SaveDataMountResult) uintptr {
	if mount2 == nil {
		logger.Printf("%-132s %s failed due to invalid mount pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount2"),
		)
		return 0x809F0000
	}
	if mount2.UserId < 0 {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount2"),
		)
		return 0x809F0011
	}
	if mount2.DirName == nil {
		logger.Printf("%-132s %s failed due to invalid dir name.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount2"),
		)
		return 0x809F0000
	}

	return saveDataMount(mount2.To1(), mountResult)
}

func saveDataMount(mount *SaveDataMount, mountResult *SaveDataMountResult) uintptr {
	create := mount.MountMode&SaveDataMountModeCreate != 0
	createIfNotExists := mount.MountMode&SaveDataMountModeCreate2 != 0
	copyIcon := mount.MountMode&SaveDataMountModeCopyIcon != 0

	// Get next available mount slot.
	mountSlot, err := GlobalSaveDataManager.GetAvailableMountSlot(GoString(mount.DirName))
	if err != nil {
		logger.Printf("%-132s %s failed due to slot error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("saveDataMount"),
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
			color.Magenta.Sprint("saveDataMount"),
		)
		return 0x809F0008
	}
	if create && exists {
		logger.Printf("%-132s %s failed due to existing save (with create flag).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("saveDataMount"),
		)
		return 0x809F0007
	}

	// Mount save instance at slot (will create if it's new).
	err, created := saveInstance.Mount(mountSlot, copyIcon)
	if err != nil {
		logger.Printf("%-132s %s failed due to mount error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("saveDataMount"),
			err.Error(),
		)
		return 0x809F000B
	}

	// Set mount result.
	CString(Cstring(unsafe.Pointer(&mountResult.MountPoint)), saveInstance.MountPoint)
	if created {
		mountResult.MountStatus = SaveDataMountStatusCreated
	}
	GlobalSaveDataManager.SetSaveInstance(mountSlot, saveInstance)

	logger.Printf("%-132s %s mounted save at %s (userId=%s, titleId=%s, dirName=%s, blocks=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("saveDataMount"),
		color.Green.Sprint(saveInstance.MountPoint),
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
	if mountPointPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mount point pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataUmount"),
		)
		return 0x809F0000
	}

	// Unmount save.
	mountPoint := GoString(Cstring(mountPointPtr))
	if err := GlobalSaveDataManager.Unmount(mountPoint); err != nil {
		logger.Printf("%-132s %s failed due to unmount error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataUmount"),
			color.Green.Sprint(mountPoint),
			err.Error(),
		)
		return 0x809F0008
	}

	logger.Printf("%-132s %s unmounted save from %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataUmount"),
		color.Green.Sprint(mountPoint),
	)
	return 0
}
