//go:build linux

package sys_struct

/*
	#include <signal.h>
	#include <ucontext.h>

	static int get_si_signo(siginfo_t* info) {
        return info->si_signo;
    }

	static void* get_si_addr(siginfo_t* info) {
        return info->si_addr;
    }

	static const char* get_signal_name(int sig) {
		switch(sig) {
			case SIGSEGV: return "SIGSEGV";
			case SIGBUS:  return "SIGBUS";
			case SIGILL:  return "SIGILL";
			case SIGTRAP: return "SIGTRAP";
			case SIGFPE:  return "SIGFPE ";
			case SIGABRT: return "SIGABRT";
			case SIGSYS:  return "SIGSYS";
			default:      return "UNKNOWN SIGNAL";
		}
	}
*/
import "C"

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/kernel"
)

type SIGNAL_CONTEXT struct {
	Info    *C.siginfo_t
	Context *C.ucontext_t
}

// Register numbers, should be the same always xd.
// https://sites.uclouvain.be/SystInfo/usr/include/sys/ucontext.h.html
const (
	REG_R8      = 0
	REG_R9      = 1
	REG_R10     = 2
	REG_R11     = 3
	REG_R12     = 4
	REG_R13     = 5
	REG_R14     = 6
	REG_R15     = 7
	REG_RDI     = 8
	REG_RSI     = 9
	REG_RBP     = 10
	REG_RBX     = 11
	REG_RDX     = 12
	REG_RAX     = 13
	REG_RCX     = 14
	REG_RSP     = 15
	REG_RIP     = 16
	REG_EFL     = 17
	REG_CSGSFS  = 18
	REG_ERR     = 19
	REG_TRAPNO  = 20
	REG_OLDMASK = 21
	REG_CR2     = 22
)

const (
	SIGNAL_SIGSEGV = C.SIGSEGV // ACCESS_VIOLATION equivalent
	SIGNAL_SIGBUS  = C.SIGBUS  // Same as ACCESS_VIOLATION
	SIGNAL_SIGILL  = C.SIGILL
	SIGNAL_SIGTRAP = C.SIGTRAP // SINGLE_STEP equivalent
	SIGNAL_SIGFPE  = C.SIGFPE
	SIGNAL_SIGABRT = C.SIGABRT
	SIGNAL_SIGSYS  = C.SIGSYS
)

func (ctx *SIGNAL_CONTEXT) GetCode() uintptr {
	return uintptr(C.get_si_signo(ctx.Info))
}

func (ctx *SIGNAL_CONTEXT) GetName() string {
	return C.GoString(C.get_signal_name(C.get_si_signo(ctx.Info)))
}

func (ctx *SIGNAL_CONTEXT) GetFaultAddress() uintptr {
	return uintptr(C.get_si_addr(ctx.Info))
}

func (ctx *SIGNAL_CONTEXT) GetRegister(regIndex int) uintptr {
	return uintptr(ctx.Context.uc_mcontext.gregs[regIndex])
}

func (ctx *SIGNAL_CONTEXT) SetRegister(regIndex int, value uintptr) {
	ctx.Context.uc_mcontext.gregs[regIndex] = C.greg_t(value)
}

func (ctx *SIGNAL_CONTEXT) CopyTo(uctx *kernel.Ucontext) {
	uctx.Mcontext.Rdi = uint64(ctx.GetRegister(REG_RDI))
	uctx.Mcontext.Rsi = uint64(ctx.GetRegister(REG_RSI))
	uctx.Mcontext.Rdx = uint64(ctx.GetRegister(REG_RDX))
	uctx.Mcontext.Rcx = uint64(ctx.GetRegister(REG_RCX))
	uctx.Mcontext.R8 = uint64(ctx.GetRegister(REG_R8))
	uctx.Mcontext.R9 = uint64(ctx.GetRegister(REG_R9))
	uctx.Mcontext.Rax = uint64(ctx.GetRegister(REG_RAX))
	uctx.Mcontext.Rbx = uint64(ctx.GetRegister(REG_RBX))
	uctx.Mcontext.Rbp = uint64(ctx.GetRegister(REG_RBP))
	uctx.Mcontext.R10 = uint64(ctx.GetRegister(REG_R10))
	uctx.Mcontext.R11 = uint64(ctx.GetRegister(REG_R11))
	uctx.Mcontext.R12 = uint64(ctx.GetRegister(REG_R12))
	uctx.Mcontext.R13 = uint64(ctx.GetRegister(REG_R13))
	uctx.Mcontext.R14 = uint64(ctx.GetRegister(REG_R14))
	uctx.Mcontext.R15 = uint64(ctx.GetRegister(REG_R15))
	uctx.Mcontext.Rip = uint64(ctx.GetRegister(REG_RIP))
	uctx.Mcontext.Rsp = uint64(ctx.GetRegister(REG_RSP))
	uctx.Mcontext.Rflags = uint64(ctx.GetRegister(REG_EFL))
	uctx.Mcontext.Addr = uint64(ctx.GetFaultAddress())
}

func (ctx *SIGNAL_CONTEXT) CopyFrom(uctx *kernel.Ucontext) {
	ctx.SetRegister(REG_RDI, uintptr(uctx.Mcontext.Rdi))
	ctx.SetRegister(REG_RSI, uintptr(uctx.Mcontext.Rsi))
	ctx.SetRegister(REG_RDX, uintptr(uctx.Mcontext.Rdx))
	ctx.SetRegister(REG_RCX, uintptr(uctx.Mcontext.Rcx))
	ctx.SetRegister(REG_R8, uintptr(uctx.Mcontext.R8))
	ctx.SetRegister(REG_R9, uintptr(uctx.Mcontext.R9))
	ctx.SetRegister(REG_RAX, uintptr(uctx.Mcontext.Rax))
	ctx.SetRegister(REG_RBX, uintptr(uctx.Mcontext.Rbx))
	ctx.SetRegister(REG_RBP, uintptr(uctx.Mcontext.Rbp))
	ctx.SetRegister(REG_R10, uintptr(uctx.Mcontext.R10))
	ctx.SetRegister(REG_R11, uintptr(uctx.Mcontext.R11))
	ctx.SetRegister(REG_R12, uintptr(uctx.Mcontext.R12))
	ctx.SetRegister(REG_R13, uintptr(uctx.Mcontext.R13))
	ctx.SetRegister(REG_R14, uintptr(uctx.Mcontext.R14))
	ctx.SetRegister(REG_R15, uintptr(uctx.Mcontext.R15))
	ctx.SetRegister(REG_RIP, uintptr(uctx.Mcontext.Rip))
	ctx.SetRegister(REG_RSP, uintptr(uctx.Mcontext.Rsp))
	ctx.SetRegister(REG_EFL, uintptr(uctx.Mcontext.Rflags))
}
