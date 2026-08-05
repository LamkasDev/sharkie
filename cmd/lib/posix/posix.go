package posix

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPosixStubs() {
	// Semaphore functions.
	RegisterPosixStub("sem_init", libScePosix_sem_init)
	RegisterPosixStub("sem_destroy", libScePosix_sem_destroy)
	RegisterPosixStub("sem_trywait", libScePosix_sem_trywait)
	RegisterPosixStub("sem_wait", libScePosix_sem_wait)
	RegisterPosixStub("sem_timedwait", libScePosix_sem_timedwait)
	RegisterPosixStub("sem_reltimedwait_np", libScePosix_sem_reltimedwait_np)
	RegisterPosixStub("sem_post", libScePosix_sem_post)

	// Clock functions.
	RegisterPosixStub("clock_getres", libScePosix_clock_getres)
	RegisterPosixStub("clock_gettime", libScePosix_clock_gettime)
	RegisterPosixStub("gettimeofday", libScePosix_gettimeofday)

	// Process functions.
	RegisterPosixStub("getpid", libScePosix_getpid)
	RegisterPosixStub("sleep", libScePosix_sleep)
	RegisterPosixStub("usleep", libScePosix_usleep)
	RegisterPosixStub("nanosleep", libScePosix_nanosleep)

	// Thread functions.
	RegisterPosixStub("pthread_create", libScePosix_pthread_create)
	RegisterPosixStub("pthread_create_name_np", libScePosix_pthread_create_name_np)
	RegisterPosixStub("pthread_self", libScePosix_pthread_self)
	RegisterPosixStub("pthread_equal", libScePosix_pthread_equal)
	RegisterPosixStub("pthread_detach", libScePosix_pthread_detach)
	RegisterPosixStub("pthread_join", libScePosix_pthread_join)
	RegisterPosixStub("pthread_cancel", libScePosix_pthread_cancel)
	RegisterPosixStub("pthread_once", libScePosix_pthread_once)
	RegisterPosixStub("pthread_yield", libScePosix_pthread_yield)
	RegisterPosixStub("pthread_exit", libScePosix_pthread_exit)
	RegisterPosixStub("pthread_getaffinity_np", libScePosix_pthread_getaffinity_np)
	RegisterPosixStub("pthread_setaffinity_np", libScePosix_pthread_setaffinity_np)
	RegisterPosixStub("pthread_getschedparam", libScePosix_pthread_getschedparam)
	RegisterPosixStub("pthread_setschedparam", libScePosix_pthread_setschedparam)
	RegisterPosixStub("pthread_setcancelstate", libScePosix_pthread_setcancelstate)

	// Scheduling functions.
	RegisterPosixStub("sched_yield", libScePosix_sched_yield)
	RegisterPosixStub("sched_get_priority_min", libScePosix_sched_get_priority_min)
	RegisterPosixStub("sched_get_priority_max", libScePosix_sched_get_priority_max)

	// Thread attribute functions.
	RegisterPosixStub("pthread_attr_init", libScePosix_pthread_attr_init)
	RegisterPosixStub("pthread_attr_getstacksize", libScePosix_pthread_attr_getstacksize)
	RegisterPosixStub("pthread_attr_setstacksize", libScePosix_pthread_attr_setstacksize)
	RegisterPosixStub("pthread_attr_getschedpolicy", libScePosix_pthread_attr_getschedpolicy)
	RegisterPosixStub("pthread_attr_setschedpolicy", libScePosix_pthread_attr_setschedpolicy)
	RegisterPosixStub("pthread_attr_setinheritsched", libScePosix_pthread_attr_setinheritsched)
	RegisterPosixStub("pthread_attr_getschedparam", libScePosix_pthread_attr_getschedparam)
	RegisterPosixStub("pthread_attr_setschedparam", libScePosix_pthread_attr_setschedparam)
	RegisterPosixStub("pthread_attr_setguardsize", libScePosix_pthread_attr_setguardsize)
	RegisterPosixStub("pthread_attr_setdetachstate", libScePosix_pthread_attr_setdetachstate)
	RegisterPosixStub("pthread_attr_setscope", libScePosix_pthread_attr_setscope)
	RegisterPosixStub("pthread_attr_getaffinity_np", libScePosix_pthread_attr_getaffinity_np)
	RegisterPosixStub("pthread_attr_destroy", libScePosix_pthread_attr_destroy)

	// Thread key functions.
	RegisterPosixStub("pthread_key_create", libScePosix_pthread_key_create)
	RegisterPosixStub("pthread_key_delete", libScePosix_pthread_key_delete)
	RegisterPosixStub("pthread_getspecific", libScePosix_pthread_getspecific)
	RegisterPosixStub("pthread_setspecific", libScePosix_pthread_setspecific)

	// Mutex functions.
	RegisterPosixStub("pthread_mutex_init", libScePosix_pthread_mutex_init)
	RegisterPosixStub("pthread_mutex_lock", libScePosix_pthread_mutex_lock)
	RegisterPosixStub("pthread_mutex_unlock", libScePosix_pthread_mutex_unlock)
	RegisterPosixStub("pthread_mutex_destroy", libScePosix_pthread_mutex_destroy)
	RegisterPosixStub("pthread_mutex_trylock", libScePosix_pthread_mutex_trylock)
	RegisterPosixStub("pthread_mutex_timedlock", libScePosix_pthread_mutex_timedlock)
	RegisterPosixStub("pthread_mutex_reltimedlock_np", libScePosix_pthread_mutex_reltimedlock_np)

	// Mutex attribute functions.
	RegisterPosixStub("pthread_mutexattr_init", libScePosix_pthread_mutexattr_init)
	RegisterPosixStub("pthread_mutexattr_settype", libScePosix_pthread_mutexattr_settype)
	RegisterPosixStub("pthread_mutexattr_setprotocol", libScePosix_pthread_mutexattr_setprotocol)
	RegisterPosixStub("pthread_mutexattr_destroy", libScePosix_pthread_mutexattr_destroy)

	// Cond functions.
	RegisterPosixStub("pthread_cond_init", libScePosix_pthread_cond_init)
	RegisterPosixStub("pthread_cond_destroy", libScePosix_pthread_cond_destroy)
	RegisterPosixStub("pthread_cond_broadcast", libScePosix_pthread_cond_broadcast)
	RegisterPosixStub("pthread_cond_signal", libScePosix_pthread_cond_signal)
	RegisterPosixStub("pthread_cond_wait", libScePosix_pthread_cond_wait)
	RegisterPosixStub("pthread_cond_timedwait", libScePosix_pthread_cond_timedwait)
	RegisterPosixStub("pthread_cond_reltimedwait_np", libScePosix_pthread_cond_reltimedwait_np)

	// Cond attribute functions.
	RegisterPosixStub("pthread_condattr_init", libScePosix_pthread_condattr_init)
	RegisterPosixStub("pthread_condattr_destroy", libScePosix_pthread_condattr_destroy)

	// Rwlock functions.
	RegisterPosixStub("pthread_rwlock_init", libScePosix_pthread_rwlock_init)
	RegisterPosixStub("pthread_rwlock_rdlock", libScePosix_pthread_rwlock_rdlock)
	RegisterPosixStub("pthread_rwlock_wrlock", libScePosix_pthread_rwlock_wrlock)
	RegisterPosixStub("pthread_rwlock_unlock", libScePosix_pthread_rwlock_unlock)

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
	RegisterPosixStub("fstat", libScePosix_fstat)
	RegisterPosixStub("truncate", libScePosix_truncate)
	RegisterPosixStub("truncate_0", libScePosix_truncate)
	RegisterPosixStub("ftruncate", libScePosix_ftruncate)
	RegisterPosixStub("ftruncate_0", libScePosix_ftruncate)
	RegisterPosixStub("ioctl", libScePosix_ioctl)
	RegisterPosixStub("shm_open", libScePosix_shm_open)

	// Memory functions.
	RegisterPosixStub("getpagesize", libScePosix_getpagesize)
	RegisterPosixStub("mmap", libScePosix_mmap)
	RegisterPosixStub("mmap_0", libScePosix_mmap)
	RegisterPosixStub("munmap", libScePosix_munmap)
	RegisterPosixStub("mname", libScePosix_mname)
	RegisterPosixStub("mprotect", libScePosix_mprotect)

	// Network functions.
	RegisterPosixStub("socketpair", libScePosix_socketpair)
	RegisterPosixStub("recvmsg", libScePosix_recvmsg)
	RegisterPosixStub("sendmsg", libScePosix_sendmsg)

	// Equeue/kevent functions.
	RegisterPosixStub("kevent", libScePosix_kevent)
	RegisterPosixStub("kqueue", libScePosix_kqueue)
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
