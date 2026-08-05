package save_data

import (
	"os"
	"path/filepath"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000029200
// __int64 __fastcall sceSaveDataGetParam(int, int, int, __int64, __int64)
func libSceSaveData_sceSaveDataGetParam(mountPointPtr, paramTypeVal, paramBuf, paramBufSize uintptr, gotSizePtr *uint64) uintptr {
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
		if gotSizePtr != nil {
			*gotSizePtr = uint64(unsafe.Sizeof(SaveDataParam{}))
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
		if gotSizePtr != nil {
			*gotSizePtr = uint64(n + 1)
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
		if gotSizePtr != nil {
			*gotSizePtr = 4
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
		if gotSizePtr != nil {
			*gotSizePtr = 8
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
func libSceSaveData_sceSaveDataSetParam(mountPointPtr Cstring, paramTypeVal, paramBuf, paramBufSize uintptr) uintptr {
	if mountPointPtr == nil || paramBuf == 0 || uint32(paramTypeVal) > uint32(SaveDataParamTypeUserParam) {
		logger.Printf("%-132s %s failed due to invalid pointers or paramType.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSaveDataSetParam"),
		)
		return 0x809F0000
	}
	mountPoint := GoString(mountPointPtr)
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
		param := (*SaveDataParam)(unsafe.Pointer(paramBuf))
		param.SaveToParamSfo(saveInstance.ParamSfo)
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
