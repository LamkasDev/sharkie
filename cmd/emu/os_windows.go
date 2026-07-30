//go:build windows

package emu

func GetOsThreadId() uint32 {
	return 0 // TODO: Windows thread ID
}

func Tgkill(tgid int, tid int, sig int) error {
	return nil // TODO: Windows tgkill equivalent
}

func PlatformToOrbisSignal(platformSignum int) int {
	return platformSignum
}

func OrbisToPlatformSignal(orbisSignum int) int {
	return orbisSignum
}
