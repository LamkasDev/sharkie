package lib_structs

import "time"

const TSC_FREQUENCY = uint64(1_600_000_000)
const PTC_FREQUENCY = uint64(1_600_000_000)

var TscStartTime = time.Now()

func ReadTsc() uintptr {
	elapsed := time.Since(TscStartTime)
	ticks := (uint64(elapsed.Nanoseconds()) * TSC_FREQUENCY) / 1_000_000_000
	return uintptr(ticks)
}

func ReadUptimeTsc() uintptr {
	return ReadTsc()
}
