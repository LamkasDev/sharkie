//go:build windows

package sys_struct

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// PrintContext prints the given context.
func PrintContext(ctx *CONTEXT) {
	logger.Println("Context:")
	PrintRegister("RAX", ctx.Rax)
	PrintRegister("RBX", ctx.Rbx)
	PrintRegister("RCX", ctx.Rcx)
	PrintRegister("RDX", ctx.Rdx)
	PrintRegister("RBP", ctx.Rbp)
	PrintRegister("RSI", ctx.Rsi)
	PrintRegister("RDI", ctx.Rdi)
	PrintRegister("RSP", ctx.Rsp)
	PrintRegister("R8", ctx.R8)
	PrintRegister("R9", ctx.R9)
	PrintRegister("R10", ctx.R10)
	PrintRegister("R11", ctx.R11)
	PrintRegister("R12", ctx.R12)
	PrintRegister("R13", ctx.R13)
	PrintRegister("R14", ctx.R14)
	PrintRegister("R15", ctx.R15)
	logger.Printf(
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
}

// PrintRegister prints the given register and it's value.
func PrintRegister(register string, value uint64) {
	logger.Printf(
		"  %42s = %s (%s)\n",
		color.Blue.Sprint(register),
		color.Yellow.Sprintf("0x%016X", value),
		color.Yellow.Sprintf("%d", value),
	)
}
