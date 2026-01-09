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
