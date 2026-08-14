package pad

type MouseData struct {
	Timestamp uint64
	Connected bool
	Buttons   uint32
	X         int32
	Y         int32
	Wheel     int32
	Tilt      int32
	Reserved  [8]byte
}
