package posix

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPosixStubs() {
	elf.RegisterStub("libScePosix", "pthread_setschedparam", libScePosix_stub)

	// Semaphore functions.
	elf.RegisterStub("libScePosix", "sem_post", libScePosix_sem_post)
	elf.RegisterStub("libScePosix", "sem_wait", libScePosix_sem_wait)
	elf.RegisterStub("libScePosix", "sem_timedwait", libScePosix_sem_timedwait)
	elf.RegisterStub("libScePosix", "sem_destroy", libScePosix_stub)
	elf.RegisterStub("libScePosix", "sem_init", libScePosix_sem_init)

	// Clock functions.
	elf.RegisterStub("libScePosix", "clock_gettime", libScePosix_clock_gettime)
	elf.RegisterStub("libScePosix", "clock_gettimeofday", libScePosix_clock_gettimeofday)
}

func libScePosix_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
