package posix

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPosixStubs() {
	// Semaphore functions.
	elf.RegisterStub("libScePosix", "sem_post", libScePosix_sem_post)
	elf.RegisterStub("libScePosix", "sem_wait", libScePosix_sem_wait)
	elf.RegisterStub("libScePosix", "sem_timedwait", libScePosix_sem_timedwait)
	elf.RegisterStub("libScePosix", "sem_destroy", libScePosix_stub)
	elf.RegisterStub("libScePosix", "sem_init", libScePosix_sem_init)

	// Clock functions.
	elf.RegisterStub("libScePosix", "clock_gettime", libScePosix_clock_gettime)
	elf.RegisterStub("libScePosix", "clock_gettimeofday", libScePosix_clock_gettimeofday)

	// Process functions.
	elf.RegisterStub("libScePosix", "usleep", libScePosix_usleep)
	elf.RegisterStub("libScePosix", "nanosleep", libScePosix_nanosleep)

	// Thread functions.
	elf.RegisterStub("libScePosix", "pthread_create", libScePosix_pthread_create)
	elf.RegisterStub("libScePosix", "pthread_create_name_np", libScePosix_pthread_create_name_np)
	elf.RegisterStub("libScePosix", "pthread_setschedparam", libScePosix_stub)

	// Mutex functions.
	elf.RegisterStub("libScePosix", "pthread_mutex_init", libScePosix_pthread_mutex_init)
	elf.RegisterStub("libScePosix", "pthread_mutex_lock", libScePosix_pthread_mutex_lock)
	elf.RegisterStub("libScePosix", "pthread_mutex_unlock", libScePosix_pthread_mutex_unlock)
	elf.RegisterStub("libScePosix", "pthread_mutex_destroy", libScePosix_pthread_mutex_destroy)

	// IO functions.
	elf.RegisterStub("libScePosix", "open", libScePosix_open)
	elf.RegisterStub("libScePosix", "read", libScePosix_read)
	elf.RegisterStub("libScePosix", "lseek", libScePosix_lseek)
	elf.RegisterStub("libScePosix", "close", libScePosix_close)
}

func libScePosix_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
