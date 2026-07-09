package time

type Timestamp struct {
	Seconds     int64
	Nanoseconds int64
}

type Timesec struct {
	UtcTime int64
	WestSec int32
	DstSec  int32
}

type Timezone struct {
	MinutesWest int32
	DstTime     int32
}

type Timevalue struct {
	Seconds      int64
	Microseconds int64
}

type Timeout struct {
	Microseconds uint32
}
