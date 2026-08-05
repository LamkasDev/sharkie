package save_data

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
)

type SaveDataMemoryOption uint32

const (
	SaveDataMemoryOptionNone              = SaveDataMemoryOption(0)
	SaveDataMemoryOptionSetParam          = SaveDataMemoryOption(1 << 0)
	SaveDataMemoryOptionDoubleBuffer      = SaveDataMemoryOption(1 << 1)
	SaveDataMemoryOptionUnlockLimitations = SaveDataMemoryOption(1 << 2)
)

type SaveDataMemorySetup2 struct {
	Option         SaveDataMemoryOption
	UserId         UserId
	MemorySize     uint64
	IconMemorySize uint64
	InitParam      *SaveDataParam
	InitIcon       *SaveDataIcon
	SlotId         uint32
	Reserved       [20]byte
}

type SaveDataMemorySetupResult struct {
	ExistedMemorySize uint64
	Reserved          [16]byte
}
