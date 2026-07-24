package posix

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPosixStubs() {
	// Semaphore functions.
	RegisterPosixStub("sem_post", libScePosix_sem_post)
	RegisterPosixStub("sem_wait", libScePosix_sem_wait)
	RegisterPosixStub("sem_timedwait", libScePosix_sem_timedwait)
	RegisterPosixStub("sem_destroy", libScePosix_stub)
	RegisterPosixStub("sem_init", libScePosix_sem_init)

	// Clock functions.
	RegisterPosixStub("clock_gettime", libScePosix_clock_gettime)
	RegisterPosixStub("clock_gettimeofday", libScePosix_clock_gettimeofday)

	// Process functions.
	RegisterPosixStub("usleep", libScePosix_usleep)
	RegisterPosixStub("nanosleep", libScePosix_nanosleep)

	// Thread functions.
	RegisterPosixStub("pthread_create", libScePosix_pthread_create)
	RegisterPosixStub("pthread_create_name_np", libScePosix_pthread_create_name_np)
	RegisterPosixStub("pthread_setschedparam", libScePosix_stub)

	// Mutex functions.
	RegisterPosixStub("pthread_mutex_init", libScePosix_pthread_mutex_init)
	RegisterPosixStub("pthread_mutex_lock", libScePosix_pthread_mutex_lock)
	RegisterPosixStub("pthread_mutex_unlock", libScePosix_pthread_mutex_unlock)
	RegisterPosixStub("pthread_mutex_destroy", libScePosix_pthread_mutex_destroy)

	// IO functions.
	RegisterPosixStub("open", libScePosix_open)
	RegisterPosixStub("_open", libScePosix_open)
	RegisterPosixStub("read", libScePosix_read)
	RegisterPosixStub("_read", libScePosix_read)
	RegisterPosixStub("pread", libScePosix_pread)
	RegisterPosixStub("pread_0", libScePosix_pread)
	RegisterPosixStub("write", libScePosix_write)
	RegisterPosixStub("_write", libScePosix_write)
	RegisterPosixStub("pwrite", libScePosix_pwrite)
	RegisterPosixStub("pwrite_0", libScePosix_pwrite)
	RegisterPosixStub("fcntl", libScePosix_fcntl)
	RegisterPosixStub("lseek", libScePosix_lseek)
	RegisterPosixStub("lseek_0", libScePosix_lseek)
	RegisterPosixStub("close", libScePosix_close)
	RegisterPosixStub("_close", libScePosix_close)
	RegisterPosixStub("stat", libScePosix_stat)
}

// We need these functions to be available in kernel; libraries take them from there.
func RegisterPosixStub(symbolName string, goFn any) {
	elf.RegisterStub("libScePosix", symbolName, goFn)
	elf.RegisterStub("libkernel", symbolName, goFn)
}

func libScePosix_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
