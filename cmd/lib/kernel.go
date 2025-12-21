package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

func RegisterKernelStubs() {
	// Stack smashing protection.
	// https://gcc.gnu.org/onlinedocs/gcc-4.1.2/gccint/Stack-Smashing-Protection.html
	stackChkGuard := elf.RegisterVariableStub("libkernel", "__stack_chk_guard", 8)
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(stackChkGuard.Address)), 8),
		0xDEADBEEF,
	)
	elf.RegisterStub("libkernel", "__stack_chk_fail", StackChkFail)

	// Environment variables.
	environ := elf.RegisterVariableStub("libkernel", "environ", 8)
	environList, _ := sys_struct.AllocReadWriteMemory(8)
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(environ.Address)), 8),
		uint64(environList),
	)

	// Pointer to current program name.
	progname := elf.RegisterVariableStub("libkernel", "__progname", 8)
	prognameStr, _ := sys_struct.AllocReadWriteMemory(32)
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(prognameStr)), 32),
		"eboot.bin\x00",
	)
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(progname.Address)), 8),
		uint64(prognameStr),
	)

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

	// Memory functions.
	elf.RegisterStub("libkernel", "mmap", libKernel_mmap)
	elf.RegisterStub("libkernel", "mmap_0", libKernel_mmap_0)
	elf.RegisterStub("libkernel", "sceKernelMmap", libKernel_sceKernelMmap)
	elf.RegisterStub("libkernel", "sub_1C90", libKernel_mname)
	elf.RegisterStub("libkernel", "sceKernelAllocateDirectMemory", libKernel_sceKernelAllocateDirectMemory)
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
	elf.RegisterStub("libkernel", "write", libKernel_write)
	elf.RegisterStub("libkernel", "_write", libKernel__write)
	elf.RegisterStub("libkernel", "ioctl", libKernel_ioctl)
	elf.RegisterStub("libkernel", "_ioctl", libKernel_ioctl) // TODO: this neither

	// Process functions.
	elf.RegisterStub("libkernel", "getpid", libKernel_getpid)
	elf.RegisterStub("libkernel", "sceKernelGetProcessType", libKernel_sceKernelGetProcessType)

	// Thread functions.
	elf.RegisterStub("libkernel", "pthread_mutexattr_init", libKernel_pthread_mutexattr_init)
	elf.RegisterStub("libkernel", "scePthreadAttrInit", libKernel_scePthreadAttrInit)

	// Mutex functions.
	elf.RegisterStub("libkernel", "pthread_mutexattr_init", libKernel_pthread_mutexattr_init)
	elf.RegisterStub("libkernel", "scePthreadMutexattrInit", libKernel_scePthreadMutexattrInit)
	elf.RegisterStub("libkernel", "pthread_mutexattr_settype", libKernel_pthread_mutexattr_settype)
	elf.RegisterStub("libkernel", "scePthreadMutexattrSettype", libKernel_scePthreadMutexattrSettype)
	elf.RegisterStub("libkernel", "pthread_mutex_init", libKernel_pthread_mutex_init)
	elf.RegisterStub("libkernel", "scePthreadMutexInit", libKernel_scePthreadMutexInit)
	elf.RegisterStub("libkernel", "pthread_mutexattr_destroy", libKernel_pthread_mutexattr_destroy)
	elf.RegisterStub("libkernel", "scePthreadMutexattrDestroy", libKernel_scePthreadMutexattrDestroy)
	elf.RegisterStub("libkernel", "pthread_mutex_lock", libKernel_pthread_mutex_lock)
	elf.RegisterStub("libkernel", "scePthreadMutexLock", libKernel_scePthreadMutexLock)
	elf.RegisterStub("libkernel", "pthread_mutex_unlock", libKernel_pthread_mutex_unlock)
	elf.RegisterStub("libkernel", "scePthreadMutexUnlock", libKernel_scePthreadMutexUnlock)

	// Cond functions.
	elf.RegisterStub("libkernel", "pthread_cond_broadcast", libKernel_pthread_cond_broadcast)
	elf.RegisterStub("libkernel", "scePthreadCondBroadcast", libKernel_scePthreadCondBroadcast)

	// Event flag functions.
	elf.RegisterStub("libkernel", "sceKernelCreateEventFlag", libKernel_sceKernelCreateEventFlag)

	// Module functions.
	elf.RegisterStub("libkernel", "sceKernelGetModuleInfoForUnwind", libKernel_sceKernelGetModuleInfoForUnwind)
	elf.RegisterStub("libkernel", "sub_1EB0", libKernel_sys_dynlib_get_info_ex)
}
