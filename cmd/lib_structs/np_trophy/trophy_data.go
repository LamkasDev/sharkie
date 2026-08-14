package np_trophy

import . "github.com/LamkasDev/sharkie/cmd/lib_structs/time"

type TrophyDetails struct {
	Size     uint64
	Id       uint32
	Grade    uint32
	GroupId  uint32
	Hidden   bool
	Reserved [3]byte

	Title       [128]byte
	Description [1024]byte
}

type TrophyData struct {
	Size      uint64
	Id        uint32
	Unlocked  bool
	Reserved  [3]byte
	Timestamp RtcTick
}
