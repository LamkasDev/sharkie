package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/mem"
)

func RegisterKernelStubs() {
	// Stack smashing protection.
	// https://gcc.gnu.org/onlinedocs/gcc-4.1.2/gccint/Stack-Smashing-Protection.html
	asm.Stubs[elf.GetSymbolHashIndex("libkernel", "__stack_chk_guard")] = asm.StubInfo{
		Address: mem.AllocReadWriteMemory(8),
	}
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(asm.Stubs[elf.GetSymbolHashIndex("libkernel", "__stack_chk_guard")].Address)), 8),
		0xDEADBEEF,
	)
	elf.RegisterStub("libkernel", "__stack_chk_fail", StackChkFail)

	// Environment variables.
	asm.Stubs[elf.GetSymbolHashIndex("libkernel", "environ")] = asm.StubInfo{
		Address: mem.AllocReadWriteMemory(8),
	}
	environList := mem.AllocReadWriteMemory(8)
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(asm.Stubs[elf.GetSymbolHashIndex("libkernel", "environ")].Address)), 8),
		uint64(environList),
	)

	// Pointer to current program name.
	asm.Stubs[elf.GetSymbolHashIndex("libkernel", "__progname")] = asm.StubInfo{
		Address: mem.AllocReadWriteMemory(8),
	}
	prognameStr := mem.AllocReadWriteMemory(32)
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(prognameStr)), 32),
		"eboot.bin\x00",
	)
	binary.LittleEndian.PutUint64(
		unsafe.Slice((*byte)(unsafe.Pointer(asm.Stubs[elf.GetSymbolHashIndex("libkernel", "__progname")].Address)), 8),
		uint64(prognameStr),
	)

	// Flag used by libc to control signal interrupt behavior.
	// https://www.gnu.org/software//libc/manual/2.23/html_node/Other-Safety-Remarks.html
	asm.Stubs[elf.GetSymbolHashIndex("libkernel", "_sigintr")] = asm.StubInfo{
		Address: mem.AllocReadWriteMemory(4),
	}

	// Error functions.
	elf.RegisterStub("libkernel", "_error", libKernel__error)

	// Memory functions.
	elf.RegisterStub("libkernel", "mmap", libKernel_mmap)
	elf.RegisterStub("libkernel", "mmap_0", libKernel_mmap_0)
	elf.RegisterStub("libkernel", "sceKernelMmap", libKernel_sceKernelMmap)
	elf.RegisterStub("libkernel", "sub_1C90", libKernel_mname)
	elf.RegisterStub("libkernel", "sceKernelMapNamedSystemFlexibleMemory", libKernel_sceKernelMapNamedSystemFlexibleMemory)

	// IO functions.
	elf.RegisterStub("libkernel", "open", libKernel_open)
	elf.RegisterStub("libkernel", "_open", libKernel__open)
	elf.RegisterStub("libkernel", "_write", libKernel__write)
	elf.RegisterStub("libkernel", "ioctl", libKernel_ioctl)

	// Process functions.
	elf.RegisterStub("libkernel", "getpid", libKernel_getpid)
	elf.RegisterStub("libkernel", "sceKernelGetProcessType", libKernel_sceKernelGetProcessType)

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
}
