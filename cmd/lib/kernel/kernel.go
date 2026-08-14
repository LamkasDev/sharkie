package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
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
	CString(Cstring(prognameStrAddr), "eboot.bin")
	WriteAddress(progname.Address, prognameStrAddr)

	// Flag used by libc to control signal interrupt behavior.
	// https://www.gnu.org/software//libc/manual/2.23/html_node/Other-Safety-Remarks.html
	elf.RegisterVariableStub("libkernel", "_sigintr", 4)

	// Syscall functions.
	elf.RegisterStub("libkernel", "sysctl", libKernel_sysctl)
	elf.RegisterStub("libkernel", "sysarch", libKernel_sys_sysarch)
	elf.RegisterStub("libkernel", "sub_1590", libKernel_sys_thr_self)
	elf.RegisterStub("libkernel", "rtprio_thread", libKernel_rtprio_thread)
	elf.RegisterStub("libkernel", "get_authinfo", libKernel_sys_get_authinfo)
	elf.RegisterStub("libkernel", "__sys_regmgr_call", libKernel___sys_regmgr_call)
	elf.RegisterStub("libkernel", "__sys_get_proc_type_info", libKernel___sys_get_proc_type_info)
	elf.RegisterStub("libkernel", "__tls_get_addr", libKernel___tls_get_addr)

	// Error functions.
	elf.RegisterStub("libkernel", "__error", libKernel___error)
	elf.RegisterStub("libkernel", "sceKernelError", libKernel_sceKernelError)

	// Exception functions.
	elf.RegisterStub("libkernel", "sceKernelInstallExceptionHandler", libKernel_sceKernelInstallExceptionHandler)
	elf.RegisterStub("libkernel", "sceKernelRemoveExceptionHandler", libKernel_sceKernelRemoveExceptionHandler)
	elf.RegisterStub("libkernel", "sceKernelRaiseException", libKernel_sceKernelRaiseException)
	elf.RegisterStub("libkernel", "sceKernelDebugRaiseException", libKernel_sceKernelDebugRaiseException)
	elf.RegisterStub("libkernel", "sceKernelDebugRaiseExceptionOnReleaseMode", libKernel_sceKernelDebugRaiseExceptionOnReleaseMode)

	// Memory functions.
	elf.RegisterStub("libkernel", "sceKernelMmap", libKernel_sceKernelMmap)
	elf.RegisterStub("libkernel", "sceKernelMunmap", libKernel_sceKernelMunmap)
	elf.RegisterStub("libkernel", "sceKernelAllocateMainDirectMemory", libKernel_sceKernelAllocateMainDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelAllocateDirectMemory", libKernel_sceKernelAllocateDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelMapDirectMemory", libKernel_sceKernelMapDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedDirectMemory", libKernel_sceKernelMapNamedDirectMemory)
	elf.RegisterStub("libkernel", "sceKernelGetDirectMemorySize", libKernel_sceKernelGetDirectMemorySize)
	elf.RegisterStub("libkernel", "sceKernelAvailableDirectMemorySize", libKernel_sceKernelAvailableDirectMemorySize)
	elf.RegisterStub("libkernel", "sceKernelAvailableFlexibleMemorySize", libKernel_sceKernelAvailableFlexibleMemorySize)
	elf.RegisterStub("libkernel", "sceKernelMprotect", libKernel_sceKernelMprotect)
	elf.RegisterStub("libkernel", "sceKernelMapFlexibleMemory", libKernel_sceKernelMapFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedFlexibleMemory", libKernel_sceKernelMapNamedFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelMapNamedSystemFlexibleMemory", libKernel_sceKernelMapNamedSystemFlexibleMemory)
	elf.RegisterStub("libkernel", "sceKernelSetVirtualRangeName", libKernel_sceKernelSetVirtualRangeName)
	elf.RegisterStub("libkernel", "sceKernelVirtualQuery", libKernel_sceKernelVirtualQuery)
	elf.RegisterStub("libkernel", "sceKernelDirectMemoryQuery", libKernel_sceKernelDirectMemoryQuery)
	elf.RegisterStub("libkernel", "sceKernelReserveVirtualRange", libKernel_sceKernelReserveVirtualRange)

	// IO functions.
	elf.RegisterStub("libkernel", "sceKernelOpen", libKernel_sceKernelOpen)
	elf.RegisterStub("libkernel", "sceKernelClose", libKernel_sceKernelClose)
	elf.RegisterStub("libkernel", "sceKernelRead", libKernel_sceKernelRead)
	elf.RegisterStub("libkernel", "sceKernelPread", libKernel_sceKernelPread)
	elf.RegisterStub("libkernel", "sceKernelWrite", libKernel_sceKernelWrite)
	elf.RegisterStub("libkernel", "sceKernelPwrite", libKernel_sceKernelPwrite)
	elf.RegisterStub("libkernel", "sceKernelTruncate", libKernel_sceKernelTruncate)
	elf.RegisterStub("libkernel", "sceKernelFtruncate", libKernel_sceKernelFtruncate)
	elf.RegisterStub("libkernel", "sceKernelLseek", libKernel_sceKernelLseek)
	elf.RegisterStub("libkernel", "sceKernelStat", libKernel_sceKernelStat)
	elf.RegisterStub("libkernel", "sceKernelFstat", libKernel_sceKernelFstat)
	elf.RegisterStub("libkernel", "sceKernelCheckReachability", libKernel_sceKernelCheckReachability)

	// Directory functions.
	elf.RegisterStub("libkernel", "sceKernelMkdir", libKernel_sceKernelMkdir)
	elf.RegisterStub("libkernel", "sceKernelGetdents", libKernel_sceKernelGetdents)
	elf.RegisterStub("libkernel", "sceKernelGetdirentries", libKernel_sceKernelGetdirentries)

	// Shared memory functions.
	elf.RegisterStub("libkernel", "shm_open", libKernel_shm_open)

	// Process functions.
	elf.RegisterStub("libkernel", "sceKernelGetProcessType", libKernel_sceKernelGetProcessType)
	elf.RegisterStub("libkernel", "sceKernelGetProcParam", libKernel_sceKernelGetProcParam)
	elf.RegisterStub("libkernel", "sceKernelGetCpumode", libKernel_sceKernelGetCpumode)
	elf.RegisterStub("libkernel", "sceKernelSleep", libKernel_sceKernelSleep)
	elf.RegisterStub("libkernel", "sceKernelUsleep", libKernel_sceKernelUsleep)
	elf.RegisterStub("libkernel", "sceKernelNanosleep", libKernel_sceKernelNanosleep)

	// Thread functions.
	elf.RegisterStub("libkernel", "scePthreadGetthreadid", libKernel_scePthreadGetthreadid)
	elf.RegisterStub("libkernel", "scePthreadSelf", libKernel_scePthreadSelf)
	elf.RegisterStub("libkernel", "scePthreadEqual", libKernel_scePthreadEqual)
	elf.RegisterStub("libkernel", "scePthreadCreate", libKernel_scePthreadCreate)
	elf.RegisterStub("libkernel", "scePthreadGetaffinity", libKernel_scePthreadGetaffinity)
	elf.RegisterStub("libkernel", "scePthreadSetaffinity", libKernel_scePthreadSetaffinity)
	elf.RegisterStub("libkernel", "scePthreadGetschedparam", libKernel_scePthreadGetschedparam)
	elf.RegisterStub("libkernel", "scePthreadSetschedparam", libKernel_scePthreadSetschedparam)
	elf.RegisterStub("libkernel", "scePthreadExit", libKernel_scePthreadExit)
	elf.RegisterStub("libkernel", "scePthreadJoin", libKernel_scePthreadJoin)
	elf.RegisterStub("libkernel", "scePthreadOnce", libKernel_scePthreadOnce)
	elf.RegisterStub("libkernel", "scePthreadYield", libKernel_scePthreadYield)

	// Thread attribute functions.
	elf.RegisterStub("libkernel", "scePthreadAttrInit", libKernel_scePthreadAttrInit)
	elf.RegisterStub("libkernel", "scePthreadAttrDestroy", libKernel_scePthreadAttrDestroy)
	elf.RegisterStub("libkernel", "scePthreadAttrSetstacksize", libKernel_scePthreadAttrSetstacksize)
	elf.RegisterStub("libkernel", "scePthreadAttrSetschedpolicy", libKernel_scePthreadAttrSetschedpolicy)
	elf.RegisterStub("libkernel", "scePthreadAttrSetinheritsched", libKernel_scePthreadAttrSetinheritsched)
	elf.RegisterStub("libkernel", "scePthreadAttrGetschedparam", libKernel_scePthreadAttrGetschedparam)
	elf.RegisterStub("libkernel", "scePthreadAttrSetschedparam", libKernel_scePthreadAttrSetschedparam)
	elf.RegisterStub("libkernel", "scePthreadAttrSetguardsize", libKernel_scePthreadAttrSetguardsize)
	elf.RegisterStub("libkernel", "scePthreadAttrSetdetachstate", libKernel_scePthreadAttrSetdetachstate)
	elf.RegisterStub("libkernel", "scePthreadAttrSetscope", libKernel_scePthreadAttrSetscope)
	elf.RegisterStub("libkernel", "scePthreadAttrGet", libKernel_scePthreadAttrGet)
	elf.RegisterStub("libkernel", "scePthreadAttrGetstack", libKernel_scePthreadAttrGetstack)
	elf.RegisterStub("libkernel", "scePthreadAttrGetstackaddr", libKernel_scePthreadAttrGetstackaddr)
	elf.RegisterStub("libkernel", "scePthreadAttrGetstacksize", libKernel_scePthreadAttrGetstacksize)
	elf.RegisterStub("libkernel", "scePthreadAttrGetaffinity", libKernel_scePthreadAttrGetaffinity)

	// Mutex functions.
	elf.RegisterStub("libkernel", "scePthreadMutexInit", libKernel_scePthreadMutexInit)
	elf.RegisterStub("libkernel", "scePthreadMutexDestroy", libKernel_scePthreadMutexDestroy)
	elf.RegisterStub("libkernel", "scePthreadMutexLock", libKernel_scePthreadMutexLock)
	elf.RegisterStub("libkernel", "scePthreadMutexTrylock", libKernel_scePthreadMutexTrylock)
	elf.RegisterStub("libkernel", "scePthreadMutexUnlock", libKernel_scePthreadMutexUnlock)
	elf.RegisterStub("libkernel", "scePthreadMutexTimedlock", libKernel_scePthreadMutexTimedlock)

	// Mutex attribute functions.
	elf.RegisterStub("libkernel", "scePthreadMutexattrInit", libKernel_scePthreadMutexattrInit)
	elf.RegisterStub("libkernel", "scePthreadMutexattrSettype", libKernel_scePthreadMutexattrSettype)
	elf.RegisterStub("libkernel", "scePthreadMutexattrSetprotocol", libKernel_scePthreadMutexattrSetprotocol)
	elf.RegisterStub("libkernel", "scePthreadMutexattrDestroy", libKernel_scePthreadMutexattrDestroy)

	// Semaphore functions.
	elf.RegisterStub("libkernel", "scePthreadSemInit", libKernel_scePthreadSemInit)
	elf.RegisterStub("libkernel", "scePthreadSemDestroy", libKernel_scePthreadSemDestroy)
	elf.RegisterStub("libkernel", "scePthreadSemTrywait", libKernel_scePthreadSemTrywait)
	elf.RegisterStub("libkernel", "scePthreadSemWait", libKernel_scePthreadSemWait)
	elf.RegisterStub("libkernel", "scePthreadSemTimedwait", libKernel_scePthreadSemTimedwait)
	elf.RegisterStub("libkernel", "scePthreadSemPost", libKernel_scePthreadSemPost)

	// Cond functions.
	elf.RegisterStub("libkernel", "scePthreadCondInit", libKernel_scePthreadCondInit)
	elf.RegisterStub("libkernel", "scePthreadCondDestroy", libKernel_scePthreadCondDestroy)
	elf.RegisterStub("libkernel", "scePthreadCondBroadcast", libKernel_scePthreadCondBroadcast)
	elf.RegisterStub("libkernel", "scePthreadCondSignal", libKernel_scePthreadCondSignal)
	elf.RegisterStub("libkernel", "scePthreadCondWait", libKernel_scePthreadCondWait)
	elf.RegisterStub("libkernel", "scePthreadCondTimedwait", libKernel_scePthreadCondTimedwait)

	// Cond attritube functions.
	elf.RegisterStub("libkernel", "scePthreadCondattrInit", libKernel_scePthreadCondattrInit)
	elf.RegisterStub("libkernel", "scePthreadCondattrDestroy", libKernel_scePthreadCondattrDestroy)

	// Rwlock functions.
	elf.RegisterStub("libkernel", "scePthreadRwlockInit", libKernel_scePthreadRwlockInit)
	elf.RegisterStub("libkernel", "scePthreadRwlockRdlock", libKernel_scePthreadRwlockRdlock)
	elf.RegisterStub("libkernel", "scePthreadRwlockWrlock", libKernel_scePthreadRwlockWrlock)
	elf.RegisterStub("libkernel", "scePthreadRwlockUnlock", libKernel_scePthreadRwlockUnlock)

	// Key functions.
	elf.RegisterStub("libkernel", "scePthreadKeyCreate", libKernel_scePthreadKeyCreate)
	elf.RegisterStub("libkernel", "scePthreadGetspecific", libKernel_scePthreadGetspecific)
	elf.RegisterStub("libkernel", "scePthreadSetspecific", libKernel_scePthreadSetspecific)

	// Event flag functions.
	elf.RegisterStub("libkernel", "sceKernelCreateEventFlag", libKernel_sceKernelCreateEventFlag)
	elf.RegisterStub("libkernel", "sceKernelOpenEventFlag", libKernel_sceKernelOpenEventFlag)
	elf.RegisterStub("libkernel", "sceKernelWaitEventFlag", libKernel_sceKernelWaitEventFlag)
	elf.RegisterStub("libkernel", "sceKernelPollEventFlag", libKernel_sceKernelPollEventFlag)
	elf.RegisterStub("libkernel", "sceKernelSetEventFlag", libKernel_sceKernelSetEventFlag)
	elf.RegisterStub("libkernel", "sceKernelClearEventFlag", libKernel_sceKernelClearEventFlag)
	elf.RegisterStub("libkernel", "sceKernelDeleteEventFlag", libKernel_sceKernelDeleteEventFlag)

	// Module functions.
	elf.RegisterStub("libkernel", "sceKernelGetExecutableModuleHandle", libKernel_sceKernelGetExecutableModuleHandle)
	elf.RegisterStub("libkernel", "sceKernelGetModuleInfo", libKernel_sceKernelGetModuleInfo)
	elf.RegisterStub("libkernel", "sceKernelGetModuleInfoForUnwind", libKernel_sceKernelGetModuleInfoForUnwind)
	elf.RegisterStub("libkernel", "sceKernelIsInSandbox", libKernel_sceKernelIsInSandbox)
	elf.RegisterStub("libkernel", "sceKernelGetCompiledSdkVersion", libKernel_sceKernelGetCompiledSdkVersion)
	elf.RegisterStub("libkernel", "sceKernelSetCallRecord", libKernel_sceKernelSetCallRecord)
	elf.RegisterStub("libkernel", "sceKernelLoadStartModuleForSysmodule", libKernel_sceKernelLoadStartModuleForSysmodule)
	elf.RegisterStub("libkernel", "sceKernelLoadStartModule", libKernel_sceKernelLoadStartModule)
	elf.RegisterStub("libkernel", "sceKernelGetModuleList", libKernel_sceKernelGetModuleList)
	elf.RegisterStub("libkernel", "sceKernelDlsym", libKernel_sceKernelDlsym)

	// Dynlib functions.
	elf.RegisterStub("libkernel", "sub_1EB0", libKernel_sys_dynlib_get_info_ex)
	elf.RegisterStub("libkernel", "sub_1D90", libKernel_sys_dynlib_process_needed_and_relocate)

	// App functions.
	elf.RegisterStub("libkernel", "sceKernelGetAppInfo", libKernel_sceKernelGetAppInfo)
	elf.RegisterStub("libkernel", "sceKernelGetSystemSwVersion", libKernel_sceKernelGetSystemSwVersion)
	elf.RegisterStub("libkernel", "sceKernelTitleWorkaroundIsEnabled", libKernel_sceKernelTitleWorkaroundIsEnabled)

	// TSC functions.
	elf.RegisterStub("libkernel", "sceKernelGetTscFrequency", libKernel_sceKernelGetTscFrequency)
	elf.RegisterStub("libkernel", "sceKernelReadTsc", libKernel_sceKernelReadTsc)
	elf.RegisterStub("libkernel", "sceKernelGetProcessTimeCounterFrequency", libKernel_sceKernelGetProcessTimeCounterFrequency)
	elf.RegisterStub("libkernel", "sceKernelGetProcessTimeCounter", libKernel_sceKernelGetProcessTimeCounter)

	// IPMI functions.
	elf.RegisterStub("libkernel", "ipmimgr_call", libKernel_ipmimgr_call)

	// Clock functions.
	elf.RegisterStub("libkernel", "sceKernelClockGettime", libKernel_sceKernelClockGettime)
	elf.RegisterStub("libkernel", "sceKernelGetProcessTime", libKernel_sceKernelGetProcessTime)
	elf.RegisterStub("libkernel", "sceKernelGettimeofday", libKernel_sceKernelGettimeofday)
	elf.RegisterStub("libkernel", "sceKernelConvertUtcToLocaltime", libKernel_sceKernelConvertUtcToLocaltime)
	elf.RegisterStub("libkernel", "sceKernelConvertLocaltimeToUtc", libKernel_sceKernelConvertLocaltimeToUtc)

	// Signal functions.
	elf.RegisterStub("libkernel", "sigprocmask", libKernel_sigprocmask)
	elf.RegisterStub("libkernel", "_sigprocmask", libKernel_sigprocmask)

	// Equeue functions.
	elf.RegisterStub("libkernel", "sceKernelCreateEqueue", libKernel_sceKernelCreateEqueue)
	elf.RegisterStub("libkernel", "sceKernelWaitEqueue", libKernel_sceKernelWaitEqueue)

	// Kevent functions.
	elf.RegisterStub("libkernel", "sceKernelAddUserEvent", libKernel_sceKernelAddUserEvent)
	elf.RegisterStub("libkernel", "sceKernelAddUserEventEdge", libKernel_sceKernelAddUserEventEdge)
	elf.RegisterStub("libkernel", "sceKernelTriggerUserEvent", libKernel_sceKernelTriggerUserEvent)
	elf.RegisterStub("libkernel", "sceKernelGetEventId", libKernel_sceKernelGetEventId)
	elf.RegisterStub("libkernel", "sceKernelGetEventFilter", libKernel_sceKernelGetEventFilter)
	elf.RegisterStub("libkernel", "sceKernelGetEventData", libKernel_sceKernelGetEventData)
	elf.RegisterStub("libkernel", "sceKernelGetEventUserData", libKernel_sceKernelGetEventUserData)

	// Kernel semaphore functions.
	elf.RegisterStub("libkernel", "sceKernelCreateSema", libKernel_sceKernelCreateSema)
	elf.RegisterStub("libkernel", "sceKernelOpenSema", libKernel_sceKernelOpenSema)
	elf.RegisterStub("libkernel", "sceKernelCancelSema", libKernel_sceKernelCancelSema)
	elf.RegisterStub("libkernel", "sceKernelDeleteSema", libKernel_sceKernelDeleteSema)
	elf.RegisterStub("libkernel", "sceKernelWaitSema", libKernel_sceKernelWaitSema)
	elf.RegisterStub("libkernel", "sceKernelPollSema", libKernel_sceKernelPollSema)
	elf.RegisterStub("libkernel", "sceKernelSignalSema", libKernel_sceKernelSignalSema)

	// Network functions.
	elf.RegisterStub("libkernel", "__sys_netcontrol", libKernel___sys_netcontrol)
	elf.RegisterStub("libkernel", "__sys_socketex", libKernel___sys_socketex)
	elf.RegisterStub("libkernel", "__sys_socketclose", libKernel___sys_socketclose)

	// Coredump functions.
	elf.RegisterStub("libkernel", "sceCoredumpRegisterCoredumpHandler", libKernel_sceCoredumpRegisterCoredumpHandler)
	elf.RegisterStub("libSceCoredump", "sceCoredumpRegisterCoredumpHandler", libKernel_sceCoredumpRegisterCoredumpHandler)
	elf.RegisterStub("libkernel", "sceCoredumpUnregisterCoredumpHandler", libKernel_sceCoredumpUnregisterCoredumpHandler)
	elf.RegisterStub("libSceCoredump", "sceCoredumpUnregisterCoredumpHandler", libKernel_sceCoredumpUnregisterCoredumpHandler)
}

func StackChkFail() uintptr {
	logger.Printf(
		"%-132s %s\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Red.Sprintf("stack check failed :("),
	)
	logger.Print(emu.SprintStackTrace())
	logger.CleanupAndExit()

	return 0
}
