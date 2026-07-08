package save_data

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

type OrbisSaveDataBlocks uint64

type OrbisSaveDataMountMode uint32

const (
	OrbisSaveDataMountModeRDONLY       = OrbisSaveDataMountMode(1 << 0)
	OrbisSaveDataMountModeRDWR         = OrbisSaveDataMountMode(1 << 1)
	OrbisSaveDataMountModeCREATE       = OrbisSaveDataMountMode(1 << 2)
	OrbisSaveDataMountModeDESTRUCT_OFF = OrbisSaveDataMountMode(1 << 3)
	OrbisSaveDataMountModeCOPY_ICON    = OrbisSaveDataMountMode(1 << 4)
	OrbisSaveDataMountModeCREATE2      = OrbisSaveDataMountMode(1 << 5)
)

type OrbisSaveDataMountStatus uint32

const (
	OrbisSaveDataMountStatusNOTHING = OrbisSaveDataMountStatus(0)
	OrbisSaveDataMountStatusCREATED = OrbisSaveDataMountStatus(1)
)

type OrbisSaveDataMount struct {
	UserId      int32
	_           uint32
	TitleId     Cstring
	DirName     Cstring
	Fingerprint Cstring
	Blocks      OrbisSaveDataBlocks
	MountMode   OrbisSaveDataMountMode
	Reserved    [32]byte
}

type OrbisSaveDataMountInfo struct {
	Blocks     OrbisSaveDataBlocks
	FreeBlocks OrbisSaveDataBlocks
	Reserved   [32]byte
}

type OrbisSaveDataMountResult struct {
	MountPoint     Cstring
	RequiredBlocks OrbisSaveDataBlocks
	_              uint32
	MountStatus    OrbisSaveDataMountStatus
	Reserved       [28]byte
	_              uint32
}
