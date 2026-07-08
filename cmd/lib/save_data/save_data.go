package save_data

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
)

func RegisterSaveDataStubs() {
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize", SceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataDirNameSearch", SceSaveDataDirNameSearch)
	elf.RegisterStub("libSceSaveData", "sceSaveDataMount", SceSaveDataMount)
	elf.RegisterStub("libSceSaveData", "sceSaveDataUmount", SceSaveDataUmount)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetMountInfo", SceSaveDataGetMountInfo)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetParam", SceSaveDataGetParam)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetParam", SceSaveDataSetParam)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSaveIcon", SceSaveDataSaveIcon)
	elf.RegisterStub("libSceSaveData", "sceSaveDataDelete", SceSaveDataDelete)

	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogGetResult", SceSaveDataDialogGetResult)
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogTerminate", SceSaveDataDialogTerminate)
}

func SceSaveDataInitialize(param uintptr) uintptr {
	logger.Printf("sceSaveDataInitialize called (param=%x)\n", param)
	return 0
}

func SceSaveDataDirNameSearch(param uintptr) uintptr {
	logger.Printf("sceSaveDataDirNameSearch called (param=%x)\n", param)
	return 0
}

func SceSaveDataMount(mountPtr uintptr, resultPtr uintptr) uintptr {
	logger.Printf("sceSaveDataMount called (mountPtr=%x, resultPtr=%x)\n", mountPtr, resultPtr)
	return 0
}

func SceSaveDataUmount(mountPointPtr uintptr) uintptr {
	logger.Printf("sceSaveDataUmount called (mountPointPtr=%x)\n", mountPointPtr)
	return 0
}

func SceSaveDataGetMountInfo(infoPtr uintptr) uintptr {
	logger.Printf("sceSaveDataGetMountInfo called (infoPtr=%x)\n", infoPtr)
	return 0
}

func SceSaveDataGetParam(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataGetParam called (paramPtr=%x)\n", paramPtr)
	return 0
}

func SceSaveDataSetParam(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataSetParam called (paramPtr=%x)\n", paramPtr)
	return 0
}

func SceSaveDataSaveIcon(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataSaveIcon called (paramPtr=%x)\n", paramPtr)
	return 0
}

func SceSaveDataDelete(paramPtr uintptr) uintptr {
	logger.Printf("sceSaveDataDelete called (paramPtr=%x)\n", paramPtr)
	return 0
}

func SceSaveDataDialogGetResult(resultPtr uintptr) uintptr {
	logger.Printf("sceSaveDataDialogGetResult called (resultPtr=%x)\n", resultPtr)
	return 0
}

func SceSaveDataDialogTerminate() uintptr {
	logger.Printf("sceSaveDataDialogTerminate called\n")
	return 0
}
