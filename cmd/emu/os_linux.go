//go:build linux

package emu

import (
	"syscall"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

func GetOsThreadId() uint32 {
	return uint32(syscall.Gettid())
}

func Tgkill(tgid int, tid int, sig int) error {
	return syscall.Tgkill(tgid, tid, syscall.Signal(sig))
}

func PlatformToOrbisSignal(platformSignum int) int {
	switch platformSignum {
	case sys_struct.SIGNAL_SIGSEGV:
		return 11
	case sys_struct.SIGNAL_SIGBUS:
		return 10
	case sys_struct.SIGNAL_SIGILL:
		return 4
	case sys_struct.SIGNAL_SIGTRAP:
		return 5
	case sys_struct.SIGNAL_SIGFPE:
		return 8
	case sys_struct.SIGNAL_SIGABRT:
		return 6
	case sys_struct.SIGNAL_SIGSYS:
		return 12
	case 10: // SIGUSR1
		return 30
	default:
		return platformSignum
	}
}

func OrbisToPlatformSignal(orbisSignum int) int {
	switch orbisSignum {
	case 11:
		return sys_struct.SIGNAL_SIGSEGV
	case 10:
		return sys_struct.SIGNAL_SIGBUS
	case 4:
		return sys_struct.SIGNAL_SIGILL
	case 5:
		return sys_struct.SIGNAL_SIGTRAP
	case 8:
		return sys_struct.SIGNAL_SIGFPE
	case 6:
		return sys_struct.SIGNAL_SIGABRT
	case 12:
		return sys_struct.SIGNAL_SIGSYS
	case 30:
		return 10 // SIGUSR1
	default:
		return orbisSignum
	}
}
