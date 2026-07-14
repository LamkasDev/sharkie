package save_data

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

type OrbisSaveDataSortKey uint32

const (
	OrbisSaveDataSortKeyDIRNAME     = OrbisSaveDataSortKey(0)
	OrbisSaveDataSortKeyUSER_PARAM  = OrbisSaveDataSortKey(1)
	OrbisSaveDataSortKeyBLOCKS      = OrbisSaveDataSortKey(2)
	OrbisSaveDataSortKeyMTIME       = OrbisSaveDataSortKey(3)
	OrbisSaveDataSortKeyFREE_BLOCKS = OrbisSaveDataSortKey(5)
)

type OrbisSaveDataSortOrder uint32

const (
	OrbisSaveDataSortOrderASCENT  = OrbisSaveDataSortOrder(0)
	OrbisSaveDataSortOrderDESCENT = OrbisSaveDataSortOrder(1)
)

type OrbisSaveDataParam struct {
	Title     [128]byte
	SubTitle  [128]byte
	Detail    [1024]byte
	UserParam uint32
	_         uint32
	Mtime     int64
	Reserved  [32]byte
}

type OrbisSaveDataDirNameSearchCond struct {
	UserId   int32
	_        uint32
	TitleId  Cstring
	DirName  Cstring
	Key      OrbisSaveDataSortKey
	Order    OrbisSaveDataSortOrder
	Reserved [32]byte
}

type OrbisSaveDataDirName struct {
	Data [32]byte
}

type OrbisSaveDataSearchInfo struct {
	Blocks     uint64
	FreeBlocks uint64
	Reserved   [32]byte
}

type OrbisSaveDataDirNameSearchResult struct {
	HitNum      uint32
	_           uint32
	DirNames    uintptr // *OrbisSaveDataDirName
	DirNamesNum uint32
	SetNum      uint32
	Params      uintptr // *OrbisSaveDataParam
	Infos       uintptr // *OrbisSaveDataSearchInfo
	Reserved    [12]byte
	_           uint32
}

type OrbisSaveDataParamType uint32

const (
	OrbisSaveDataParamTypeALL        = OrbisSaveDataParamType(0)
	OrbisSaveDataParamTypeTITLE      = OrbisSaveDataParamType(1)
	OrbisSaveDataParamTypeSUB_TITLE  = OrbisSaveDataParamType(2)
	OrbisSaveDataParamTypeDETAIL     = OrbisSaveDataParamType(3)
	OrbisSaveDataParamTypeUSER_PARAM = OrbisSaveDataParamType(4)
	OrbisSaveDataParamTypeMTIME      = OrbisSaveDataParamType(5)
)
