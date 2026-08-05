package save_data

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/config"
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

// 0x00000000000292C0
// __int64 __fastcall sceSaveDataSaveIcon(int, int)
func libSceSaveData_sceSaveDataSaveIcon(mountPointPtr Cstring, icon *SaveDataIcon) uintptr {
	if mountPointPtr == nil || icon == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSaveIcon"),
		)
		return 0x809F0000
	}
	if icon.Buf == 0 {
		return 0x809F0000
	}
	mountPoint := GoString(mountPointPtr)

	// Find save instance.
	var saveInstance *SaveInstance
	for _, save := range GlobalSaveDataManager.MountSlots {
		if save != nil && save.MountPoint == mountPoint && save.Mounted() {
			saveInstance = save
			break
		}
	}
	if saveInstance == nil {
		logger.Printf("%-132s %s failed finding save instance.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSaveIcon"),
		)
		return 0x809F000B
	}

	// Write icon.
	iconPath := filepath.Join(config.GetGameSaveDir(saveInstance.DirName), "sce_sys", "save_data.png")
	size := icon.BufSize
	if icon.DataSize < size {
		size = icon.DataSize
	}
	bufSlice := unsafe.Slice((*byte)(unsafe.Pointer(icon.Buf)), size)
	if err := os.WriteFile(iconPath, bufSlice, 0755); err != nil {
		logger.Printf("%-132s %s failed due to write error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSaveIcon"),
			err.Error(),
		)
		return 0x809F0001
	}

	logger.Printf("%-132s %s saved icon to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataSaveIcon"),
		color.Green.Sprint(iconPath),
	)
	return 0
}

// 0x0000000000028CD0
// __int64 __fastcall sceSaveDataDelete(int)
func libSceSaveData_sceSaveDataDelete(del *SaveDataDelete) uintptr {
	if del == nil {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDelete"),
		)
		return 0x809F0000
	}
	if del.DirName == nil {
		logger.Printf("%-132s %s failed due to invalid directory name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDelete"),
		)
		return 0x809F0000
	}
	dirName := strings.TrimRight(GoString(del.DirName), "\x00")
	if dirName == "" {
		logger.Printf("%-132s %s failed due to invalid directory name.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDelete"),
		)
		return 0x809F0000
	}

	// Check if mounted.
	for _, save := range GlobalSaveDataManager.MountSlots {
		if save != nil && save.DirName == dirName && save.Mounted() {
			logger.Printf("%-132s %s failed deleting mounted save instance.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataDelete"),
			)
			return 0x809F0015
		}
	}

	// Delete on host.
	savePath := config.GetGameSaveDir(dirName)
	if _, err := os.Stat(savePath); err == nil {
		if err = os.RemoveAll(savePath); err != nil {
			logger.Printf("%-132s %s failed due to remove error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataDelete"),
				err.Error(),
			)
			return 0x809F0001
		}
	}

	logger.Printf("%-132s %s deleted save %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataDelete"),
		color.Green.Sprint(dirName),
	)
	return 0
}

// 0x00000000000278E0
// __int64 sceSaveDataTerminate()
func libSceSaveData_sceSaveDataTerminate() uintptr {
	for _, save := range GlobalSaveDataManager.MountSlots {
		if save != nil {
			logger.Printf("%-132s %s failed due to mounted slot.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataTerminate"),
			)
			return 0x809F0003
		}
	}

	logger.Printf("%-132s %s terminated.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataTerminate"),
	)
	return 0
}
