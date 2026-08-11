package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitSOPP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	case gcnSpec.SoppOpWaitcnt:
		// No-op in SPIR-V for now.
	case gcnSpec.SoppOpBranch:
		// No-op in SPIR-V; CFG handler automatically emits the branch.
	case gcnSpec.SoppOpCbranchVccz:
		valLo, valHi := ctx.GetOperand64Value(b, gcnSpec.OpVccLo, 0)
		val64 := ctx.Pack64(b, valLo, valHi)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val64, ctx.GetConstId(ConstId64Uint0))
	case gcnSpec.SoppOpCbranchVccnz:
		valLo, valHi := ctx.GetOperand64Value(b, gcnSpec.OpVccLo, 0)
		val64 := ctx.Pack64(b, valLo, valHi)
		ctx.GcnConditionId = b.EmitINotEqual(ctx.GetId(BlockContextIdTypeBool), val64, ctx.GetConstId(ConstId64Uint0))
	case gcnSpec.SoppOpCbranchExecz:
		valLo, valHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		val64 := ctx.Pack64(b, valLo, valHi)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val64, ctx.GetConstId(ConstId64Uint0))
	case gcnSpec.SoppOpCbranchScc0:
		val32 := ctx.GetOperandUintValue(b, gcnSpec.OpScc, 0)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val32, ctx.GetConstId(ConstIdUint0))
	case gcnSpec.SoppOpCbranchScc1:
		val32 := ctx.GetOperandUintValue(b, gcnSpec.OpScc, 0)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val32, ctx.GetConstId(ConstIdUint1))
	case gcnSpec.SoppOpBarrier:
		b.EmitControlBarrier(
			b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), spec.SpvScopeWorkgroup),
			b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), spec.SpvScopeWorkgroup),
			b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), spec.SpvMemorySemanticsAcquireRelease|spec.SpvMemorySemanticsWorkgroupMemory))
	case gcnSpec.SoppOpEndpgm:
		// Not sure about this lol.
	case gcnSpec.SoppOpNop:
		// Nop.
	default:
		panic(fmt.Sprintf("unknown sopp op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPP][details.Op]))
	}
}
