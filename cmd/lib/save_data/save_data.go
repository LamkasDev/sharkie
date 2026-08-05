package save_data

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterSaveDataStubs() {
	// Setup functions.
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize2", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize3", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSaveIcon", libSceSaveData_sceSaveDataSaveIcon)
	elf.RegisterStub("libSceSaveData", "sceSaveDataDelete", libSceSaveData_sceSaveDataDelete)
	elf.RegisterStub("libSceSaveData", "sceSaveDataTerminate", libSceSaveData_sceSaveDataTerminate)

	// Mount functions.
	elf.RegisterStub("libSceSaveData", "sceSaveDataMount", libSceSaveData_sceSaveDataMount)
	elf.RegisterStub("libSceSaveData", "sceSaveDataMount2", libSceSaveData_sceSaveDataMount2)
	elf.RegisterStub("libSceSaveData", "sceSaveDataUmount", libSceSaveData_sceSaveDataUmount)

	// Memory functions.
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetupSaveDataMemory", libSceSaveData_sceSaveDataSetupSaveDataMemory)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetupSaveDataMemory2", libSceSaveData_sceSaveDataSetupSaveDataMemory2)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetSaveDataMemory2", libSceSaveData_sceSaveDataGetSaveDataMemory2)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetSaveDataMemory2", libSceSaveData_sceSaveDataSetSaveDataMemory2)

	// Info functions.
	elf.RegisterStub("libSceSaveData", "sceSaveDataDirNameSearch", libSceSaveData_sceSaveDataDirNameSearch)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetMountInfo", libSceSaveData_sceSaveDataGetMountInfo)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetProgress", libSceSaveData_sceSaveDataGetProgress)

	// Param functions.
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetParam", libSceSaveData_sceSaveDataGetParam)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetParam", libSceSaveData_sceSaveDataSetParam)
}
