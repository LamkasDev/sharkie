package structs

type Timestamp struct {
	Seconds     uint64
	Nanoseconds uint64
}

type Timevalue struct {
	Seconds      uint64
	Microseconds uint64
}

type Timeout struct {
	Microseconds uint32
}
