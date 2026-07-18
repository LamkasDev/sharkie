package save_data

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
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
func libSceSaveData_sceSaveDataDirNameSearch(condPtr, resultPtr uintptr) uintptr {
	if condPtr == 0 || resultPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDirNameSearch"),
		)
		return 0x809F0000
	}
	cond := (*SaveDataDirNameSearchCond)(unsafe.Pointer(condPtr))
	result := (*SaveDataDirNameSearchResult)(unsafe.Pointer(resultPtr))

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
		if result.DirNames != 0 {
			dirNames := (*[1024]SaveDataDirName)(unsafe.Pointer(result.DirNames))
			copy(dirNames[i].Data[:], make([]byte, 32))
			copy(dirNames[i].Data[:], res.DirName)
		}
		if result.Params != 0 {
			params := (*[1024]SaveDataParam)(unsafe.Pointer(result.Params))
			copy(params[i].Title[:], make([]byte, 128))
			copy(params[i].Subtitle[:], make([]byte, 128))
			copy(params[i].Detail[:], make([]byte, 1024))

			copy(params[i].Title[:], res.Psf.MapStrings[SaveParamMainTitle])
			copy(params[i].Subtitle[:], res.Psf.MapStrings[SaveParamSubtitle])
			copy(params[i].Detail[:], res.Psf.MapStrings[SaveParamDetail])
			params[i].UserParam = uint32(res.Psf.MapIntegers[SaveParamSaveDataListParam])
			params[i].MTime = res.SfoInfo.ModTime().Unix()
		}
		if result.Infos != 0 {
			infos := (*[1024]SaveDataSearchInfo)(unsafe.Pointer(result.Infos))
			infos[i].Blocks = GetMaxBlocksFromSfo(res.Psf)
			infos[i].FreeBlocks = infos[i].Blocks
		}
	}

	logger.Printf("%-132s %s found %d save dirs.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataDirNameSearch"),
		len(searchResults),
	)
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

	mount := (*SaveDataMount)(unsafe.Pointer(mountPtr))
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
	mountResult := (*SaveDataMountResult)(unsafe.Pointer(resultPtr))

	return saveDataMount(mount, mountResult)
}

// 0x0000000000027C50
// __int64 __fastcall sceSaveDataMount2(__int64, int, __m128 _XMM0)
func libSceSaveData_sceSaveDataMount2(mountPtr, resultPtr uintptr) uintptr {
	if mountPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mount pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataMount2"),
		)
		return 0x809F0000
	}

	mount2 := (*SaveDataMount2)(unsafe.Pointer(mountPtr))
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
	mount := mount2.To1()
	mountResult := (*SaveDataMountResult)(unsafe.Pointer(resultPtr))

	return saveDataMount(mount, mountResult)
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

// 0x0000000000028C20
// __int64 __fastcall sceSaveDataGetMountInfo(int, int)
func libSceSaveData_sceSaveDataGetMountInfo(mountPointPtr, infoPtr uintptr) uintptr {
	if mountPointPtr == 0 || infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataGetMountInfo"),
		)
		return 0x809F0000
	}
	mountPoint := GoString(Cstring(mountPointPtr))

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
	info := (*SaveDataMountInfo)(unsafe.Pointer(infoPtr))
	info.Blocks = SaveDataBlocks(saveInstance.MaxBlocks)
	info.FreeBlocks = info.Blocks

	logger.Printf("%-132s %s got mount info for %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataGetMountInfo"),
		color.Green.Sprint(mountPoint),
	)
	return 0
}

// 0x0000000000029200
// __int64 __fastcall sceSaveDataGetParam(int, int, int, __int64, __int64)
func libSceSaveData_sceSaveDataGetParam(mountPointPtr, paramTypeVal, paramBuf, paramBufSize, gotSizePtr uintptr) uintptr {
	if mountPointPtr == 0 || paramBuf == 0 || uint32(paramTypeVal) > uint32(SaveDataParamTypeMTime) {
		logger.Printf("%-132s %s failed due to invalid pointers or paramType.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataGetParam"),
		)
		return 0x809F0000
	}
	mountPoint := GoString(Cstring(mountPointPtr))
	paramType := SaveDataParamType(paramTypeVal)

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
			color.Magenta.Sprint("sceSaveDataGetParam"),
		)
		return 0x809F000B
	}

	// Read parameter.
	switch paramType {
	case SaveDataParamTypeAll:
		if paramBufSize != unsafe.Sizeof(SaveDataParam{}) {
			logger.Printf("%-132s %s failed due to parameter buffer size mismatch.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataGetParam"),
			)
			return 0x809F0000
		}
		params := (*SaveDataParam)(unsafe.Pointer(paramBuf))
		copy(params.Title[:], saveInstance.ParamSfo.MapStrings[SaveParamMainTitle])
		copy(params.Subtitle[:], saveInstance.ParamSfo.MapStrings[SaveParamSubtitle])
		copy(params.Detail[:], saveInstance.ParamSfo.MapStrings[SaveParamDetail])
		params.UserParam = uint32(saveInstance.ParamSfo.MapIntegers[SaveParamSaveDataListParam])
		sfoPath := filepath.Join(config.GetGameSaveDir(saveInstance.DirName), "sce_sys", "param.sfo")
		if sfoInfo, err := os.Stat(sfoPath); err == nil {
			params.MTime = sfoInfo.ModTime().Unix()
		}
		if gotSizePtr != 0 {
			*(*uint64)(unsafe.Pointer(gotSizePtr)) = uint64(unsafe.Sizeof(SaveDataParam{}))
		}
	case SaveDataParamTypeTitle, SaveDataParamTypeSubtitle, SaveDataParamTypeDetail:
		key := ""
		if paramType == SaveDataParamTypeTitle {
			key = SaveParamMainTitle
		} else if paramType == SaveDataParamTypeSubtitle {
			key = SaveParamSubtitle
		} else if paramType == SaveDataParamTypeDetail {
			key = SaveParamDetail
		}
		val := saveInstance.ParamSfo.MapStrings[key]
		bufSlice := unsafe.Slice((*byte)(unsafe.Pointer(paramBuf)), paramBufSize)
		n := copy(bufSlice, val)
		if n < int(paramBufSize) {
			bufSlice[n] = 0 // null terminate.
		} else {
			bufSlice[paramBufSize-1] = 0
			n = int(paramBufSize - 1)
		}
		if gotSizePtr != 0 {
			*(*uint64)(unsafe.Pointer(gotSizePtr)) = uint64(n + 1)
		}
	case SaveDataParamTypeUserParam:
		if paramBufSize < 4 {
			logger.Printf("%-132s %s failed due to parameter buffer size mismatch.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataGetParam"),
			)
			return 0x809F0000
		}
		*(*uint32)(unsafe.Pointer(paramBuf)) = uint32(saveInstance.ParamSfo.MapIntegers[SaveParamSaveDataListParam])
		if gotSizePtr != 0 {
			*(*uint64)(unsafe.Pointer(gotSizePtr)) = 4
		}
	case SaveDataParamTypeMTime:
		if paramBufSize < 8 {
			logger.Printf("%-132s %s failed due to parameter buffer size mismatch.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataGetParam"),
			)
			return 0x809F0000
		}
		sfoPath := filepath.Join(config.GetGameSaveDir(saveInstance.DirName), "sce_sys", "param.sfo")
		if sfoInfo, err := os.Stat(sfoPath); err == nil {
			*(*int64)(unsafe.Pointer(paramBuf)) = sfoInfo.ModTime().Unix()
		} else {
			*(*int64)(unsafe.Pointer(paramBuf)) = 0
		}
		if gotSizePtr != 0 {
			*(*uint64)(unsafe.Pointer(gotSizePtr)) = 8
		}
	}

	logger.Printf("%-132s %s read paramType %d.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataGetParam"),
		paramType,
	)
	return 0
}

// 0x0000000000029140
// __int64 __fastcall sceSaveDataSetParam(int, int, int, __int64)
func libSceSaveData_sceSaveDataSetParam(mountPointPtr, paramTypeVal, paramBuf, paramBufSize uintptr) uintptr {
	if mountPointPtr == 0 || paramBuf == 0 || uint32(paramTypeVal) > uint32(SaveDataParamTypeUserParam) {
		logger.Printf("%-132s %s failed due to invalid pointers or paramType.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSetParam"),
		)
		return 0x809F0000
	}
	mountPoint := GoString(Cstring(mountPointPtr))
	paramType := SaveDataParamType(paramTypeVal)

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
			color.Magenta.Sprint("sceSaveDataSetParam"),
		)
		return 0x809F000B
	}

	// Set parameter.
	switch paramType {
	case SaveDataParamTypeAll:
		if paramBufSize != unsafe.Sizeof(SaveDataParam{}) {
			logger.Printf("%-132s %s failed due to parameter buffer size mismatch.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataSetParam"),
			)
			return 0x809F0000
		}
		params := (*SaveDataParam)(unsafe.Pointer(paramBuf))
		titleIdx := bytes.IndexByte(params.Title[:], 0)
		if titleIdx == -1 {
			titleIdx = len(params.Title)
		}
		saveInstance.ParamSfo.MapStrings[SaveParamMainTitle] = string(params.Title[:titleIdx])
		subTitleIdx := bytes.IndexByte(params.Subtitle[:], 0)
		if subTitleIdx == -1 {
			subTitleIdx = len(params.Subtitle)
		}
		saveInstance.ParamSfo.MapStrings[SaveParamSubtitle] = string(params.Subtitle[:subTitleIdx])
		detailIdx := bytes.IndexByte(params.Detail[:], 0)
		if detailIdx == -1 {
			detailIdx = len(params.Detail)
		}
		saveInstance.ParamSfo.MapStrings[SaveParamDetail] = string(params.Detail[:detailIdx])
		saveInstance.ParamSfo.MapIntegers[SaveParamSaveDataListParam] = int32(params.UserParam)
	case SaveDataParamTypeTitle, SaveDataParamTypeSubtitle, SaveDataParamTypeDetail:
		key := ""
		if paramType == SaveDataParamTypeTitle {
			key = SaveParamMainTitle
		} else if paramType == SaveDataParamTypeSubtitle {
			key = SaveParamSubtitle
		} else if paramType == SaveDataParamTypeDetail {
			key = SaveParamDetail
		}
		val := GoString(Cstring(paramBuf))
		saveInstance.ParamSfo.MapStrings[key] = val
	case SaveDataParamTypeUserParam:
		if paramBufSize < 4 {
			logger.Printf("%-132s %s failed due to parameter buffer size mismatch.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceSaveDataSetParam"),
			)
			return 0x809F0000
		}
		val := *(*int32)(unsafe.Pointer(paramBuf))
		saveInstance.ParamSfo.MapIntegers[SaveParamSaveDataListParam] = val
	}

	logger.Printf("%-132s %s set paramType %d.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSaveDataSetParam"),
		paramType,
	)
	return 0
}

// 0x00000000000292C0
// __int64 __fastcall sceSaveDataSaveIcon(int, int)
func libSceSaveData_sceSaveDataSaveIcon(mountPointPtr, iconPtr uintptr) uintptr {
	if mountPointPtr == 0 || iconPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSaveIcon"),
		)
		return 0x809F0000
	}
	icon := (*SaveDataIcon)(unsafe.Pointer(iconPtr))
	if icon.Buf == 0 {
		return 0x809F0000
	}
	mountPoint := GoString(Cstring(mountPointPtr))

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
func libSceSaveData_sceSaveDataDelete(delPtr uintptr) uintptr {
	if delPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataDelete"),
		)
		return 0x809F0000
	}
	del := (*SaveDataDelete)(unsafe.Pointer(delPtr))
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
