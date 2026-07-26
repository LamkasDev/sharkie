package save_data

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

type SaveDataDirNameSearchCond struct {
	UserId   int32
	_        uint32
	TitleId  Cstring
	DirName  Cstring
	Key      SaveDataSortKey
	Order    SaveDataSortOrder
	Reserved [32]byte
}

type SaveDataDirName struct {
	Data [32]byte
}

type SaveDataSearchInfo struct {
	Blocks     uint64
	FreeBlocks uint64
	Reserved   [32]byte
}

type SaveDataDirNameSearchResult struct {
	HitNum      uint32
	_           uint32
	DirNames    *[1024]SaveDataDirName
	DirNamesNum uint32
	SetNum      uint32
	Params      *[1024]SaveDataParam
	Infos       *[1024]SaveDataSearchInfo
	Reserved    [12]byte
	_           uint32
}
