package lib

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	. "github.com/LamkasDev/sharkie/cmd/structs"
)

func RegisterKernelStubs() {
	// Stack smashing protection.
	// https://gcc.gnu.org/onlinedocs/gcc-4.1.2/gccint/Stack-Smashing-Protection.html
	stackChkGuard := elf.RegisterVariableStub("libkernel", "__stack_chk_guard", 8)
	WriteAddress(stackChkGuard.Address, 0xDEADBEEF)
	elf.RegisterStub("libkernel", "__stack_chk_fail", StackChkFail)

	// Environment variables.
	environ := elf.RegisterVariableStub("libkernel", "environ", 8)
	environListAddr := GlobalGoAllocator.Malloc(8)
	WriteAddress(environ.Address, environListAddr)

	// Pointer to current program name.
	progname := elf.RegisterVariableStub("libkernel", "__progname", 8)
	prognameStrAddr := GlobalGoAllocator.Malloc(32)
	WriteCString(prognameStrAddr, "eboot.bin")
	WriteAddress(progname.Address, prognameStrAddr)

	// Flag used by libc to control signal interrupt behavior.
	// https://www.gnu.org/software//libc/manual/2.23/html_node/Other-Safety-Remarks.html
	elf.RegisterVariableStub("libkernel", "_sigintr", 4)

	// Syscall functions.
	elf.RegisterStub("libkernel", "sysctl", libKernel_sysctl)
	elf.RegisterStub("libkernel", "sysarch", libKernel_sys_sysarch)
	elf.RegisterStub("libkernel", "sub_1590", libKernel_sys_thr_self)
	elf.RegisterStub("libkernel", "rtprio_thread", libKernel_rtprio_thread)
	elf.RegisterStub("libkernel", "sub_2BA0", libKernel_sys_umtx_op)
	elf.RegisterStub("libkernel", "get_authinfo", libKernel_sys_get_authinfo)
	elf.RegisterStub("libkernel", "__sys_regmgr_call", libKernel___sys_regmgr_call)
	elf.RegisterStub("libkernel", "__sys_get_proc_type_info", libKernel___sys_get_proc_type_info)
	elf.RegisterStub("libkernel", "__tls_get_addr", libKernel___tls_get_addr)

	// Error functions.
	elf.RegisterStub("libkernel", "__error", libKernel___error)
	elf.RegisterStub("libkernel", "sceKernelError", libKernel_sceKernelError)
	elf.RegisterStub("libkernel", "sceKernelDebugRaiseException", libKernel_sceKernelDebugRaiseException)

	// Memory functions.
	elf.RegisterStub("libkernel", "mmap", libKernel_mmap)
	elf.RegisterStub("libkernel", "mmap_0", libKernel_mmap_0)
	elf.RegisterStub("libkernel", "sceKernelMmap", libKernel_sceKernelMmap)
	elf.RegisterStub("libkernel", "munmap", libKernel_munmap)
	elf.RegisterStub("libkernel", "sceKernelMunmap", libKernel_sceKernelMunmap)
	elf.RegisterStub("libkernel", "sub_1C90", libKernel_mname)
	elf.RegisterStub("libkernel", "sceKernelAllocateDirectMemory", libKernel_sceKernelAllocateDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelMapDirectMemory", libKernel_sceKernelMapDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedDirectMemory", libKernel_sceKernelMapNamedDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelGetDirectMemorySize", libKernel_sceKernelGetDirectMemorySize)
	elf.RegisterStub("libkernel", "sceKernelMprotect", libKernel_sceKernelMprotect)
	elf.RegisterStub("libkernel", "sceKernelMapFlexibleMemory", libKernel_sceKernelMapFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedFlexibleMemory", libKernel_sceKernelMapNamedFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedSystemFlexibleMemory", libKernel_sceKernelMapNamedSystemFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelSetVirtualRangeName", libKernel_sceKernelSetVirtualRangeName)

	// TODO: i have no idea what this is, it's not anywhere.
	elf.RegisterStub("libSceLibcInternal", "GG6441JdYkA#A#B", libKernel_fake)

	// IO functions.
	elf.RegisterStub("libkernel", "open", libKernel_open)
	elf.RegisterStub("libkernel", "_open", libKernel__open)
	elf.RegisterStub("libkernel", "sceKernelOpen", libKernel_sceKernelOpen)
	elf.RegisterStub("libkernel", "close", libKernel_close)
	elf.RegisterStub("libkernel", "_close", libKernel__close)
	elf.RegisterStub("libkernel", "sceKernelClose", libKernel_sceKernelClose)
	elf.RegisterStub("libkernel", "read", libKernel_read)
	elf.RegisterStub("libkernel", "_read", libKernel__read)
	elf.RegisterStub("libkernel", "sceKernelRead", libKernel_sceKernelRead)
	elf.RegisterStub("libkernel", "pread_0", libKernel_pread_0)
	elf.RegisterStub("libkernel", "pread", libKernel_pread)
	elf.RegisterStub("libkernel", "sceKernelPread", libKernel_sceKernelPread)
	elf.RegisterStub("libkernel", "write", libKernel_write)
	elf.RegisterStub("libkernel", "_write", libKernel__write)
	elf.RegisterStub("libkernel", "sceKernelWrite", libKernel_sceKernelWrite)
	elf.RegisterStub("libkernel", "pwrite_0", libKernel_pwrite_0)
	elf.RegisterStub("libkernel", "pwrite", libKernel_pwrite)
	elf.RegisterStub("libkernel", "sceKernelPwrite", libKernel_sceKernelPwrite)
	elf.RegisterStub("libkernel", "ioctl", libKernel_ioctl)
	elf.RegisterStub("libkernel", "_ioctl", libKernel_ioctl)
	elf.RegisterStub("libkernel", "truncate", libKernel_truncate)
	elf.RegisterStub("libkernel", "truncate_0", libKernel_truncate_0)
	elf.RegisterStub("libkernel", "sceKernelTruncate", libKernel_sceKernelTruncate)
	elf.RegisterStub("libkernel", "ftruncate", libKernel_ftruncate)
	elf.RegisterStub("libkernel", "ftruncate_0", libKernel_ftruncate_0)
	elf.RegisterStub("libkernel", "sceKernelFtruncate", libKernel_sceKernelFtruncate)
	elf.RegisterStub("libkernel", "stat", libKernel_stat)
	elf.RegisterStub("libkernel", "sceKernelStat", libKernel_sceKernelStat)
	elf.RegisterStub("libkernel", "fstat", libKernel_fstat)
	elf.RegisterStub("libkernel", "sceKernelFstat", libKernel_sceKernelFstat)

	// Shared memory functions.
	elf.RegisterStub("libkernel", "shm_open", libKernel_shm_open)

	// Process functions.
	elf.RegisterStub("libkernel", "getpid", libKernel_getpid)
	elf.RegisterStub("libkernel", "sceKernelGetProcessType", libKernel_sceKernelGetProcessType)
	elf.RegisterStub("libkernel", "sceKernelGetProcParam", libKernel_sceKernelGetProcParam)
	elf.RegisterStub("libkernel", "sceKernelUsleep", libKernel_sceKernelUsleep)
	elf.RegisterStub("libkernel", "sceKernelNanosleep", libKernel_sceKernelNanosleep)

	// Thread functions.
	elf.RegisterStub("libkernel", "pthread_mutexattr_init", libKernel_pthread_mutexattr_init)
	elf.RegisterStub("libkernel", "scePthreadAttrInit", libKernel_scePthreadAttrInit)
	elf.RegisterStub("libkernel", "pthread_attr_destroy", libKernel_pthread_attr_destroy)
	elf.RegisterStub("libkernel", "scePthreadAttrDestroy", libKernel_scePthreadAttrDestroy)
	elf.RegisterStub("libkernel", "pthread_attr_setstacksize", libKernel_pthread_attr_setstacksize)
	elf.RegisterStub("libkernel", "scePthreadAttrSetstacksize", libKernel_scePthreadAttrSetstacksize)
	elf.RegisterStub("libkernel", "pthread_attr_setschedpolicy", libKernel_pthread_attr_setschedpolicy)
	elf.RegisterStub("libkernel", "scePthreadAttrSetschedpolicy", libKernel_scePthreadAttrSetschedpolicy)
	elf.RegisterStub("libkernel", "pthread_attr_setinheritsched", libKernel_pthread_attr_setinheritsched)
	elf.RegisterStub("libkernel", "scePthreadAttrSetinheritsched", libKernel_scePthreadAttrSetinheritsched)
	elf.RegisterStub("libkernel", "pthread_attr_setschedparam", libKernel_pthread_attr_setschedparam)
	elf.RegisterStub("libkernel", "scePthreadAttrSetschedparam", libKernel_scePthreadAttrSetschedparam)
	elf.RegisterStub("libkernel", "pthread_attr_setguardsize", libKernel_pthread_attr_setguardsize)
	elf.RegisterStub("libkernel", "scePthreadAttrSetguardsize", libKernel_scePthreadAttrSetguardsize)
	elf.RegisterStub("libkernel", "pthread_attr_setdetachstate", libKernel_pthread_attr_setdetachstate)
	elf.RegisterStub("libkernel", "scePthreadAttrSetdetachstate", libKernel_scePthreadAttrSetdetachstate)
	elf.RegisterStub("libkernel", "pthread_attr_setscope", libKernel_pthread_attr_setscope)
	elf.RegisterStub("libkernel", "scePthreadAttrSetscope", libKernel_scePthreadAttrSetscope)
	elf.RegisterStub("libkernel", "scePthreadAttrGet", libKernel_scePthreadAttrGet)
	elf.RegisterStub("libkernel", "scePthreadAttrGetstack", libKernel_scePthreadAttrGetstack)
	elf.RegisterStub("libkernel", "pthread_attr_getaffinity_np", libKernel_pthread_attr_getaffinity_np)
	elf.RegisterStub("libkernel", "scePthreadAttrGetaffinity", libKernel_scePthreadAttrGetaffinity)
	elf.RegisterStub("libkernel", "scePthreadGetthreadid", libKernel_scePthreadGetthreadid)
	elf.RegisterStub("libkernel", "pthread_self", libKernel_pthread_self)
	elf.RegisterStub("libkernel", "scePthreadSelf", libKernel_scePthreadSelf)
	elf.RegisterStub("libkernel", "pthread_equal", libKernel_pthread_equal)
	elf.RegisterStub("libkernel", "scePthreadEqual", libKernel_scePthreadEqual)
	elf.RegisterStub("libkernel", "pthread_create_name_np", libKernel_pthread_create_name_np)
	elf.RegisterStub("libkernel", "scePthreadCreate", libKernel_scePthreadCreate)
	elf.RegisterStub("libkernel", "pthread_getaffinity_np", libKernel_pthread_getaffinity_np)
	elf.RegisterStub("libkernel", "scePthreadGetaffinity", libKernel_scePthreadGetaffinity)
	elf.RegisterStub("libkernel", "pthread_setaffinity_np", libKernel_pthread_setaffinity_np)
	elf.RegisterStub("libkernel", "scePthreadSetaffinity", libKernel_scePthreadSetaffinity)
	elf.RegisterStub("libkernel", "pthread_exit", libKernel_pthread_exit)
	elf.RegisterStub("libkernel", "scePthreadExit", libKernel_scePthreadExit)
	elf.RegisterStub("libkernel", "scePthreadRwlockRdlock", libKernel_scePthreadRwlockRdlock)
	elf.RegisterStub("libkernel", "scePthreadRwlockWrlock", libKernel_scePthreadRwlockWrlock)
	elf.RegisterStub("libkernel", "scePthreadRwlockUnlock", libKernel_scePthreadRwlockUnlock)

	// Mutex functions.
	elf.RegisterStub("libkernel", "pthread_mutexattr_init", libKernel_pthread_mutexattr_init)
	elf.RegisterStub("libkernel", "scePthreadMutexattrInit", libKernel_scePthreadMutexattrInit)
	elf.RegisterStub("libkernel", "pthread_mutexattr_destroy", libKernel_pthread_mutexattr_destroy)
	elf.RegisterStub("libkernel", "scePthreadMutexattrDestroy", libKernel_scePthreadMutexattrDestroy)
	elf.RegisterStub("libkernel", "pthread_mutexattr_settype", libKernel_pthread_mutexattr_settype)
	elf.RegisterStub("libkernel", "scePthreadMutexattrSettype", libKernel_scePthreadMutexattrSettype)
	elf.RegisterStub("libkernel", "pthread_mutex_init", libKernel_pthread_mutex_init)
	elf.RegisterStub("libkernel", "scePthreadMutexInit", libKernel_scePthreadMutexInit)
	elf.RegisterStub("libkernel", "pthread_mutex_destroy", libKernel_pthread_mutex_destroy)
	elf.RegisterStub("libkernel", "scePthreadMutexDestroy", libKernel_scePthreadMutexDestroy)
	elf.RegisterStub("libkernel", "pthread_mutex_lock", libKernel_pthread_mutex_lock)
	elf.RegisterStub("libkernel", "scePthreadMutexLock", libKernel_scePthreadMutexLock)
	elf.RegisterStub("libkernel", "pthread_mutex_trylock", libKernel_pthread_mutex_trylock)
	elf.RegisterStub("libkernel", "scePthreadMutexTrylock", libKernel_scePthreadMutexTrylock)
	elf.RegisterStub("libkernel", "pthread_mutex_unlock", libKernel_pthread_mutex_unlock)
	elf.RegisterStub("libkernel", "scePthreadMutexUnlock", libKernel_scePthreadMutexUnlock)
	elf.RegisterStub("libkernel", "pthread_mutex_timedlock", libKernel_pthread_mutex_timedlock)
	elf.RegisterStub("libkernel", "pthread_mutex_reltimedlock_np", libKernel_pthread_mutex_reltimedlock_np)
	elf.RegisterStub("libkernel", "scePthreadMutexTimedlock", libKernel_scePthreadMutexTimedlock)

	// Cond functions.
	elf.RegisterStub("libkernel", "pthread_cond_init", libKernel_pthread_cond_init)
	elf.RegisterStub("libkernel", "scePthreadCondInit", libKernel_scePthreadCondInit)
	elf.RegisterStub("libkernel", "pthread_cond_destroy", libKernel_pthread_cond_destroy)
	elf.RegisterStub("libkernel", "scePthreadCondDestroy", libKernel_scePthreadCondDestroy)
	elf.RegisterStub("libkernel", "pthread_cond_broadcast", libKernel_pthread_cond_broadcast)
	elf.RegisterStub("libkernel", "scePthreadCondBroadcast", libKernel_scePthreadCondBroadcast)
	elf.RegisterStub("libkernel", "pthread_cond_signal", libKernel_pthread_cond_signal)
	elf.RegisterStub("libkernel", "scePthreadCondSignal", libKernel_scePthreadCondSignal)
	elf.RegisterStub("libkernel", "pthread_cond_wait", libKernel_pthread_cond_wait)
	elf.RegisterStub("libkernel", "scePthreadCondWait", libKernel_scePthreadCondWait)
	elf.RegisterStub("libkernel", "pthread_cond_timedwait", libKernel_pthread_cond_timedwait)
	elf.RegisterStub("libkernel", "pthread_cond_reltimedwait_np", libKernel_pthread_cond_reltimedwait_np)
	elf.RegisterStub("libkernel", "scePthreadCondTimedwait", libKernel_scePthreadCondTimedwait)

	// Event flag functions.
	elf.RegisterStub("libkernel", "sceKernelCreateEventFlag", libKernel_sceKernelCreateEventFlag)
	elf.RegisterStub("libkernel", "sceKernelOpenEventFlag", libKernel_sceKernelOpenEventFlag)
	elf.RegisterStub("libkernel", "sceKernelWaitEventFlag", libKernel_sceKernelWaitEventFlag)
	elf.RegisterStub("libkernel", "sceKernelPollEventFlag", libKernel_sceKernelPollEventFlag)
	elf.RegisterStub("libkernel", "sceKernelSetEventFlag", libKernel_sceKernelSetEventFlag)

	// Module functions.
	elf.RegisterStub("libkernel", "sceKernelGetExecutableModuleHandle", libKernel_sceKernelGetExecutableModuleHandle)
	elf.RegisterStub("libkernel", "sceKernelGetModuleInfo", libKernel_sceKernelGetModuleInfo)
	elf.RegisterStub("libkernel", "sceKernelGetModuleInfoForUnwind", libKernel_sceKernelGetModuleInfoForUnwind)
	elf.RegisterStub("libkernel", "sub_1EB0", libKernel_sys_dynlib_get_info_ex)
	elf.RegisterStub("libkernel", "sceKernelIsInSandbox", libKernel_sceKernelIsInSandbox)
	elf.RegisterStub("libkernel", "sceKernelGetCompiledSdkVersion", libKernel_sceKernelGetCompiledSdkVersion)
	elf.RegisterStub("libkernel", "sceKernelLoadStartModuleForSysmodule", libKernel_sceKernelLoadStartModuleForSysmodule)
	elf.RegisterStub("libkernel", "sceKernelLoadStartModule", libKernel_sceKernelLoadStartModule)
	elf.RegisterStub("libkernel", "sub_1D90", libKernel_sys_dynlib_process_needed_and_relocate)

	// App functions.
	elf.RegisterStub("libkernel", "sceKernelGetAppInfo", libKernel_sceKernelGetAppInfo)
	elf.RegisterStub("libkernel", "sceKernelTitleWorkaroundIsEnabled", libKernel_sceKernelTitleWorkaroundIsEnabled)

	// TSC functions.
	elf.RegisterStub("libkernel", "sceKernelGetTscFrequency", libKernel_sceKernelGetTscFrequency)
	elf.RegisterStub("libkernel", "sceKernelReadTsc", libKernel_sceKernelReadTsc)

	// IPMI functions.
	elf.RegisterStub("libkernel", "ipmimgr_call", libKernel_ipmimgr_call)

	// Clock functions.
	elf.RegisterStub("libkernel", "clock_gettime", libKernel_clock_gettime)
	elf.RegisterStub("libkernel", "sceKernelClockGettime", libKernel_sceKernelClockGettime)
	elf.RegisterStub("libkernel", "sceKernelGetProcessTime", libKernel_sceKernelGetProcessTime)
	elf.RegisterStub("libkernel", "sceKernelGettimeofday", libKernel_sceKernelGettimeofday)

	// Signal functions.
	elf.RegisterStub("libkernel", "sigprocmask", libKernel_sigprocmask)
	elf.RegisterStub("libkernel", "_sigprocmask", libKernel_sigprocmask)

	// Equeue/kevent functions.
	elf.RegisterStub("libkernel", "kevent", libKernel_kevent)
	elf.RegisterStub("libkernel", "__sys_kqueueex", libKernel___sys_kqueueex)
	elf.RegisterStub("libkernel", "kqueue", libKernel_kqueue)
	elf.RegisterStub("libkernel", "sceKernelCreateEqueue", libKernel_sceKernelCreateEqueue)
	elf.RegisterStub("libkernel", "sceKernelWaitEqueue", libKernel_sceKernelWaitEqueue)
	elf.RegisterStub("libkernel", "sceKernelAddUserEvent", libKernel_sceKernelAddUserEvent)

	// Semaphore functions.
	elf.RegisterStub("libkernel", "sem_init", libKernel_sem_init)
	elf.RegisterStub("libkernel", "sceKernelCreateSema", libKernel_sceKernelCreateSema)
	elf.RegisterStub("libkernel", "sceKernelOpenSema", libKernel_sceKernelOpenSema)
	elf.RegisterStub("libkernel", "sceKernelDeleteSema", libKernel_sceKernelDeleteSema)
	elf.RegisterStub("libkernel", "sceKernelWaitSema", libKernel_sceKernelWaitSema)
	elf.RegisterStub("libkernel", "sem_wait", libKernel_sem_wait)
	elf.RegisterStub("libkernel", "sem_timedwait", libKernel_sem_timedwait)
	elf.RegisterStub("libkernel", "sem_post", libKernel_sem_post)

	// Network functions.
	elf.RegisterStub("libkernel", "__sys_netcontrol", libKernel___sys_netcontrol)
	elf.RegisterStub("libkernel", "__sys_socketex", libKernel___sys_socketex)
	elf.RegisterStub("libkernel", "__sys_socketclose", libKernel___sys_socketclose)
}
