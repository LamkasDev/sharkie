package save_data

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/app_content"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000002BA20
// __int64 __fastcall sceSaveDataSetupSaveDataMemory(unsigned int, __int64, __int64)
func libSceSaveData_sceSaveDataSetupSaveDataMemory(userId UserId, memorySize uintptr, param *SaveDataParam) uintptr {
	return 0
}

// 0x000000000002BC30
// __int64 __fastcall sceSaveDataSetupSaveDataMemory2(void *, __int64)
func libSceSaveData_sceSaveDataSetupSaveDataMemory2(setup2 *SaveDataMemorySetup2, result *SaveDataMemorySetupResult) uintptr {
	if setup2 == nil {
		logger.Printf("%-132s %s failed due to invalid setup pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSetupSaveDataMemory2"),
		)
		return 0x809F0000
	}

	// Prepare parameters.
	titleId, ok := GlobalAppContentInstance.ParamSfo.GetString("TITLE_ID")
	if !ok {
		panic("missing title id")
	}
	dirName := fmt.Sprintf("sce_sdmemory%d", setup2.SlotId)

	// Check if slot is not occupied.
	if _, err := GlobalSaveDataManager.GetAvailableMountSlot(dirName); err != nil {
		logger.Printf("%-132s %s failed due to slot being occupied.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSetupSaveDataMemory2"),
		)
		return 0x809F0003
	}

	// Load initial icon.
	var err error
	var initIconData []byte
	if setup2.InitIcon != nil {
		initIconData = unsafe.Slice((*byte)(unsafe.Pointer(setup2.InitIcon.Buf)), setup2.InitIcon.BufSize)
	} else if GlobalFilesystem.Exists("/app0/sce_sys/save_data.png") {
		if initIconData, err = GlobalFilesystem.ReadFull("/app0/sce_sys/save_data.png"); err != nil {
			logger.Printf("%-132s %s failed due to read error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataSetupSaveDataMemory2"),
				err.Error(),
			)
			return 0x809F000B
		}
	}

	// Create a new save memory.
	saveMemory := NewSaveMemory(setup2.UserId, titleId, dirName, setup2.MemorySize)

	// Load save memory (will create if it's new).
	err, _ = saveMemory.Load(initIconData, setup2.InitParam)
	if err != nil {
		logger.Printf("%-132s %s failed due to load error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSetupSaveDataMemory2"),
			err.Error(),
		)
		return 0x809F000B
	}

	// Set setup result.
	if result != nil {
		result.ExistedMemorySize = uint64(len(saveMemory.Data))
	}

	logger.Printf("%-132s %s setup save memory at slot %s (userId=%s, memory_size=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataSetupSaveDataMemory2"),
		color.Green.Sprint(setup2.SlotId),
		color.Yellow.Sprintf("0x%X", setup2.UserId),
		color.Yellow.Sprintf("0x%X", setup2.MemorySize),
	)
	return 0
}

// 0x000000000002C020
// __int64 __fastcall sceSaveDataGetSaveDataMemory2(__int64)
func libSceSaveData_sceSaveDataGetSaveDataMemory2() uintptr {
	return 0
}

// 0x000000000002C300
// __int64 __fastcall sceSaveDataSetSaveDataMemory2(__int64)
func libSceSaveData_sceSaveDataSetSaveDataMemory2() uintptr {
	return 0
}
