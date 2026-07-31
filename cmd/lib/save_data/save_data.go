package save_data

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterSaveDataStubs() {
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize2", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataInitialize3", libSceSaveData_sceSaveDataInitialize)
	elf.RegisterStub("libSceSaveData", "sceSaveDataDirNameSearch", libSceSaveData_sceSaveDataDirNameSearch)
	elf.RegisterStub("libSceSaveData", "sceSaveDataMount", libSceSaveData_sceSaveDataMount)
	elf.RegisterStub("libSceSaveData", "sceSaveDataMount2", libSceSaveData_sceSaveDataMount2)
	elf.RegisterStub("libSceSaveData", "sceSaveDataUmount", libSceSaveData_sceSaveDataUmount)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetMountInfo", libSceSaveData_sceSaveDataGetMountInfo)
	elf.RegisterStub("libSceSaveData", "sceSaveDataGetParam", libSceSaveData_sceSaveDataGetParam)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSetParam", libSceSaveData_sceSaveDataSetParam)
	elf.RegisterStub("libSceSaveData", "sceSaveDataSaveIcon", libSceSaveData_sceSaveDataSaveIcon)
	elf.RegisterStub("libSceSaveData", "sceSaveDataDelete", libSceSaveData_sceSaveDataDelete)
}
