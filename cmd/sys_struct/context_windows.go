//go:build windows

package sys_struct

import (
	"fmt"

	"github.com/gookit/color"
)

// SprintContext prints the given context.
func SprintContext(ctx *CONTEXT) (result string) {
	tcbAddress, _, _ := TlsGetValue.Call(PlaystationTlsSlot)
	result = "Context:\n"
	result += SprintRegister("RAX", ctx.Rax)
	result += SprintRegister("RBX", ctx.Rbx)
	result += SprintRegister("RCX", ctx.Rcx)
	result += SprintRegister("RDX", ctx.Rdx)
	result += SprintRegister("RBP", ctx.Rbp)
	result += SprintRegister("RSI", ctx.Rsi)
	result += SprintRegister("RDI", ctx.Rdi)
	result += SprintRegister("RSP", ctx.Rsp)
	result += SprintRegister("R8", ctx.R8)
	result += SprintRegister("R9", ctx.R9)
	result += SprintRegister("R10", ctx.R10)
	result += SprintRegister("R11", ctx.R11)
	result += SprintRegister("R12", ctx.R12)
	result += SprintRegister("R13", ctx.R13)
	result += SprintRegister("R14", ctx.R14)
	result += SprintRegister("R15", ctx.R15)
	result += fmt.Sprintf(
		"  %42s = [%s = %s, %s = %s, %s = %s, %s = %s, %s = %s, %s = %s]\n",
		color.Blue.Sprint("Segments"),
		color.Blue.Sprint("CS"),
		color.Yellow.Sprintf("%d", ctx.SegCs),
		color.Blue.Sprint("DS"),
		color.Yellow.Sprintf("%d", ctx.SegDs),
		color.Blue.Sprint("ES"),
		color.Yellow.Sprintf("%d", ctx.SegEs),
		color.Blue.Sprint("FS"),
		color.Yellow.Sprintf("%d", ctx.SegFs),
		color.Blue.Sprint("GS"),
		color.Yellow.Sprintf("%d", ctx.SegGs),
		color.Blue.Sprint("SS"),
		color.Yellow.Sprintf("%d", ctx.SegSs),
	)
	result += SprintRegister("TCB", uint64(tcbAddress))

	return result
}

// SprintRegister prints the given register and it's value.
func SprintRegister(register string, value uint64) string {
	return fmt.Sprintf(
		"  %42s = %s (%s)\n",
		color.Blue.Sprint(register),
		color.Yellow.Sprintf("0x%016X", value),
		color.Yellow.Sprintf("%d", value),
	)
}
