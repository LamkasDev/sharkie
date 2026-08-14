package np_trophy

type TrophyGameDetails struct {
	Size        uint64
	NumGroups   uint32
	NumTrophies uint32
	NumPlatinum uint32
	NumGold     uint32
	NumSilver   uint32
	NumBronze   uint32

	Title       [128]byte
	Description [1024]byte
}

type TrophyGameData struct {
	Size               uint64
	UnlockedTrophies   uint32
	UnlockedPlatinum   uint32
	UnlockedGold       uint32
	UnlockedSilver     uint32
	UnlockedBronze     uint32
	ProgressPercentage uint32
}
