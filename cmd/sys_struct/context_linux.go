//go:build linux

package sys_struct

// SprintContext prints the given context.
func SprintContext(ctx *SIGNAL_CONTEXT) (result string) {
	result = "Context:\n"
	result += SprintRegister("RAX", uint64(ctx.Context.uc_mcontext.gregs[REG_RAX]))
	result += SprintRegister("RBX", uint64(ctx.Context.uc_mcontext.gregs[REG_RBX]))
	result += SprintRegister("RCX", uint64(ctx.Context.uc_mcontext.gregs[REG_RCX]))
	result += SprintRegister("RDX", uint64(ctx.Context.uc_mcontext.gregs[REG_RDX]))
	result += SprintRegister("RBP", uint64(ctx.Context.uc_mcontext.gregs[REG_RBP]))
	result += SprintRegister("RSI", uint64(ctx.Context.uc_mcontext.gregs[REG_RSI]))
	result += SprintRegister("RDI", uint64(ctx.Context.uc_mcontext.gregs[REG_RDI]))
	result += SprintRegister("RSP", uint64(ctx.Context.uc_mcontext.gregs[REG_RSP]))
	result += SprintRegister("R8", uint64(ctx.Context.uc_mcontext.gregs[REG_R8]))
	result += SprintRegister("R9", uint64(ctx.Context.uc_mcontext.gregs[REG_R9]))
	result += SprintRegister("R10", uint64(ctx.Context.uc_mcontext.gregs[REG_R10]))
	result += SprintRegister("R11", uint64(ctx.Context.uc_mcontext.gregs[REG_R11]))
	result += SprintRegister("R12", uint64(ctx.Context.uc_mcontext.gregs[REG_R12]))
	result += SprintRegister("R13", uint64(ctx.Context.uc_mcontext.gregs[REG_R13]))
	result += SprintRegister("R14", uint64(ctx.Context.uc_mcontext.gregs[REG_R14]))
	result += SprintRegister("R15", uint64(ctx.Context.uc_mcontext.gregs[REG_R15]))
	result += SprintRegister("Segments (CS:GS:FS)", uint64(ctx.Context.uc_mcontext.gregs[REG_CSGSFS]))

	return result
}
