package save_data

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000028F00
// __int64 __fastcall sceSaveDataDirNameSearch(int, int)
func libSceSaveData_sceSaveDataDirNameSearch(cond *SaveDataDirNameSearchCond, result *SaveDataDirNameSearchResult) uintptr {
	if cond == nil || result == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDirNameSearch"),
		)
		return 0x809F0000
	}

	// Fetch save directories.
	savesDir := config.GetGameSavesDir()
	entries, err := os.ReadDir(savesDir)
	if err != nil {
		result.HitNum = 0
		result.SetNum = 0
		logger.Printf("%-132s %s failed due to read dir error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount"),
			err.Error(),
		)
		return 0 // TODO: error?
	}
	filterStr := ""
	if cond.DirName != nil {
		filterStr = strings.ToLower(strings.TrimRight(GoString(cond.DirName), "\x00"))
	}

	// Filter saves matching filters.
	type SaveSearchResult struct {
		DirName string
		SfoPath string
		SfoInfo os.FileInfo
		Psf     *psf.Psf
	}
	var searchResults []SaveSearchResult
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		if filterStr != "" && !strings.Contains(strings.ToLower(dirName), filterStr) {
			continue
		}
		sfoPath := filepath.Join(config.GetGameSaveDir(dirName), "sce_sys", "param.sfo")
		sfoInfo, err := os.Stat(sfoPath)
		if err != nil {
			continue
		}
		p, err := psf.NewPsfFromPath(sfoPath)
		if err != nil {
			continue
		}
		searchResults = append(searchResults, SaveSearchResult{
			DirName: dirName,
			SfoPath: sfoPath,
			SfoInfo: sfoInfo,
			Psf:     p,
		})
	}
	sort.SliceStable(searchResults, func(i, j int) bool {
		a := searchResults[i]
		b := searchResults[j]
		var less bool
		switch cond.Key {
		case SaveDataSortKeyDirName:
			less = a.DirName < b.DirName
		case SaveDataSortKeyUserParam:
			less = a.Psf.MapIntegers[SaveParamSaveDataListParam] < b.Psf.MapIntegers[SaveParamSaveDataListParam]
		case SaveDataSortKeyBlocks:
			less = GetMaxBlocksFromSfo(a.Psf) < GetMaxBlocksFromSfo(b.Psf)
		case SaveDataSortKeyFreeBlocks:
			less = GetMaxBlocksFromSfo(a.Psf) < GetMaxBlocksFromSfo(b.Psf)
		case SaveDataSortKeyMTime:
			less = a.SfoInfo.ModTime().Before(b.SfoInfo.ModTime())
		}
		if cond.Order == SaveDataSortOrderDescent {
			return !less
		}
		return less
	})

	// Populate result.
	maxCount := len(searchResults)
	if result.DirNamesNum > 0 && uint32(maxCount) > result.DirNamesNum {
		maxCount = int(result.DirNamesNum)
	}
	result.HitNum = uint32(len(searchResults))
	result.SetNum = uint32(maxCount)
	for i := 0; i < maxCount; i++ {
		res := searchResults[i]
		if result.DirNames != nil {
			copy(result.DirNames[i].Data[:], make([]byte, 32))
			copy(result.DirNames[i].Data[:], res.DirName)
		}
		if result.Params != nil {
			copy(result.Params[i].Title[:], make([]byte, 128))
			copy(result.Params[i].Subtitle[:], make([]byte, 128))
			copy(result.Params[i].Detail[:], make([]byte, 1024))

			copy(result.Params[i].Title[:], res.Psf.MapStrings[SaveParamMainTitle])
			copy(result.Params[i].Subtitle[:], res.Psf.MapStrings[SaveParamSubtitle])
			copy(result.Params[i].Detail[:], res.Psf.MapStrings[SaveParamDetail])
			result.Params[i].UserParam = uint32(res.Psf.MapIntegers[SaveParamSaveDataListParam])
			result.Params[i].MTime = res.SfoInfo.ModTime().Unix()
		}
		if result.Infos != nil {
			result.Infos[i].Blocks = GetMaxBlocksFromSfo(res.Psf)
			result.Infos[i].FreeBlocks = result.Infos[i].Blocks
		}
	}

	logger.Printf("%-132s %s found %d save dirs.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataDirNameSearch"),
		len(searchResults),
	)
	return 0
}

// 0x0000000000028C20
// __int64 __fastcall sceSaveDataGetMountInfo(int, int)
func libSceSaveData_sceSaveDataGetMountInfo(mountPointPtr Cstring, info *SaveDataMountInfo) uintptr {
	if mountPointPtr == nil || info == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataGetMountInfo"),
		)
		return 0x809F0000
	}
	mountPoint := GoString(mountPointPtr)

	// Find mounted save instance.
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
			color.Magenta.Sprint("sceSaveDataGetMountInfo"),
		)
		return 0x809F000B
	}

	// Populate result.
	info.Blocks = SaveDataBlocks(saveInstance.MaxBlocks)
	info.FreeBlocks = info.Blocks

	logger.Printf("%-132s %s got mount info for %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataGetMountInfo"),
		color.Green.Sprint(mountPoint),
	)
	return 0
}

// 0x000000000002CCA0
// __int64 __fastcall sceSaveDataGetProgress(__int64)
func libSceSaveData_sceSaveDataGetProgress(progress *float32) uintptr {
	if progress == nil {
		logger.Printf("%-132s %s failed due to invalid progress pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataGetProgress"),
		)
		return 0x809F0000
	}

	// TODO: finish this
	backupProgress := float32(1.00)
	*progress = backupProgress

	logger.Printf("%-132s %s returned backup progress %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataGetProgress"),
		color.Green.Sprint(backupProgress),
	)
	return 0
}
