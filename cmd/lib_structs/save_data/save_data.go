package save_data

import "math"

const (
	SaveDataBlockSize = uint64(32768)
	SaveDataBlocksMin = uint64(96)
	SaveDataBlocksMax = math.MaxUint64
)

type SaveDataSortKey uint32

const (
	SaveDataSortKeyDirName    = SaveDataSortKey(0)
	SaveDataSortKeyUserParam  = SaveDataSortKey(1)
	SaveDataSortKeyBlocks     = SaveDataSortKey(2)
	SaveDataSortKeyMTime      = SaveDataSortKey(3)
	SaveDataSortKeyFreeBlocks = SaveDataSortKey(5)
)

type SaveDataSortOrder uint32

const (
	SaveDataSortOrderAscent  = SaveDataSortOrder(0)
	SaveDataSortOrderDescent = SaveDataSortOrder(1)
)

type SaveDataParamType uint32

const (
	SaveDataParamTypeAll       = SaveDataParamType(0)
	SaveDataParamTypeTitle     = SaveDataParamType(1)
	SaveDataParamTypeSubtitle  = SaveDataParamType(2)
	SaveDataParamTypeDetail    = SaveDataParamType(3)
	SaveDataParamTypeUserParam = SaveDataParamType(4)
	SaveDataParamTypeMTime     = SaveDataParamType(5)
)

type SaveDataParam struct {
	Title     [128]byte
	Subtitle  [128]byte
	Detail    [1024]byte
	UserParam uint32
	_         uint32
	MTime     int64
	Reserved  [32]byte
}
