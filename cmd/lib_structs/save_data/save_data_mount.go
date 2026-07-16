package save_data

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

type SaveDataBlocks uint64

type SaveDataMountMode uint32

const (
	SaveDataMountModeReadOnly    = SaveDataMountMode(1 << 0)
	SaveDataMountModeReadWrite   = SaveDataMountMode(1 << 1)
	SaveDataMountModeCreate      = SaveDataMountMode(1 << 2)
	SaveDataMountModeDestructOff = SaveDataMountMode(1 << 3)
	SaveDataMountModeCopyIcon    = SaveDataMountMode(1 << 4)
	SaveDataMountModeCreate2     = SaveDataMountMode(1 << 5)
)

type SaveDataMountStatus uint32

const (
	SaveDataMountStatusNothing = SaveDataMountStatus(0)
	SaveDataMountStatusCreated = SaveDataMountStatus(1)
)

type SaveDataFingerprint struct {
	Data [65]byte
	_    [15]byte
}

type SaveDataMountPoint [16]byte

type SaveDataMount struct {
	UserId      int32
	_           uint32
	TitleId     Cstring
	DirName     Cstring
	Fingerprint *SaveDataFingerprint
	Blocks      SaveDataBlocks
	MountMode   SaveDataMountMode
	Reserved    [32]byte
}

type SaveDataMountInfo struct {
	Blocks     SaveDataBlocks
	FreeBlocks SaveDataBlocks
	Reserved   [32]byte
}

type SaveDataMountResult struct {
	MountPoint     SaveDataMountPoint
	RequiredBlocks SaveDataBlocks
	_              uint32
	MountStatus    SaveDataMountStatus
	Reserved       [28]byte
	_              uint32
}

type SaveDataIcon struct {
	Buf      uintptr
	BufSize  uint64
	DataSize uint64
	Reserved [32]byte
}

type SaveDataDelete struct {
	UserId   int32
	_        uint32
	TitleId  Cstring
	DirName  Cstring
	Unused   uint32
	Reserved [32]byte
	_        uint32
}
