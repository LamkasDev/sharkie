package time

type ClockId uint32

const (
	ClockIdRealtime         = ClockId(0)
	ClockIdVirtual          = ClockId(1)
	ClockIdProf             = ClockId(2)
	ClockIdMonotonic        = ClockId(4)
	ClockIdUptime           = ClockId(5)
	ClockIdUptimePrecise    = ClockId(7)
	ClockIdUptimeFast       = ClockId(8)
	ClockIdRealtimePrecise  = ClockId(9)
	ClockIdRealtimeFast     = ClockId(10)
	ClockIdMonotonicPrecise = ClockId(11)
	ClockIdMonotonicFast    = ClockId(12)
	ClockIdSecond           = ClockId(13)
	ClockIdThreadCputimeId  = ClockId(14)
	ClockIdProctime         = ClockId(15)
	ClockIdExtNetwork       = ClockId(16)
	ClockIdExtDebugNetwork  = ClockId(17)
	ClockIdExtAdNetwork     = ClockId(18)
	ClockIdExtRawNetwork    = ClockId(19)
)
