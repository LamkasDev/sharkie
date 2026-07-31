package common

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

type SpirvBlockContext struct {
	Stage    GcnShaderStage
	Address  uintptr
	LabelIds []SpirvId
	Ids      map[SpirvId]SpirvUsedId
	ConstIds map[SpirvId]SpirvUsedId

	GcnSgprArrayId SpirvId
	GcnVgprArrayId SpirvId
	GcnSpecialIds  [27]SpirvUsedId
	GcnConstIds    [120]SpirvUsedId
	GcnConditionId SpirvId
	StaticLayout   []ShaderResourceBinding

	// SpirvShaderContext contents.
	ThreadX uint32
	ThreadY uint32
	ThreadZ uint32

	PsInControl     uint32
	PsInputAddress  uint32
	PsInputControls [32]uint32

	FetchShaderAddress      uintptr
	FetchShaderInstructions []*gcnSpec.Instruction
}

type ShaderResourceBinding struct {
	InstructionOffset uintptr
	Kind              ImageAccessKind
	BindingIndex      uint32
	Access            BindingAccess
}

type BindingAccess uint8

const (
	BindingAccessSampledRead BindingAccess = iota
	BindingAccessStorageWrite
)

func (access BindingAccess) String() string {
	switch access {
	case BindingAccessSampledRead:
		return "read"
	case BindingAccessStorageWrite:
		return "write"
	}

	return "??"
}

func (ctx *SpirvBlockContext) StaticBindingIndexConst(b *SpvBuilder, instrOffset uintptr) SpirvId {
	for _, binding := range ctx.StaticLayout {
		if binding.InstructionOffset == instrOffset {
			return b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), binding.BindingIndex)
		}
	}
	panic(fmt.Sprintf("static binding not found for instruction 0x%X", instrOffset))
}

func (ctx *SpirvBlockContext) GetLabelId(i int) SpirvId {
	id := ctx.LabelIds[i]
	if id == 0 {
		panic(fmt.Sprintf("label id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetId(i SpirvId) SpirvId {
	id := ctx.Ids[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}
	id.Used = true
	ctx.Ids[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetConstId(i SpirvId) SpirvId {
	id := ctx.ConstIds[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("const id %d is zero", i))
	}
	id.Used = true
	ctx.ConstIds[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnSgprPtr(b *SpvBuilder, reg uint32) SpirvId {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnSgprArrayId, ctx.GetConstId(ConstIdUint0+SpirvId(reg)))
}

func (ctx *SpirvBlockContext) GetGcnSgprId(b *SpvBuilder, reg uint32) SpirvId {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnSgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnSgprId(b *SpvBuilder, reg uint32, value SpirvId) {
	b.EmitStore(ctx.GetGcnSgprPtr(b, reg), value)
}

func (ctx *SpirvBlockContext) GetGcnVgprPtr(b *SpvBuilder, reg uint32) SpirvId {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnVgprArrayId, ctx.GetConstId(ConstIdUint0+SpirvId(reg)))
}

func (ctx *SpirvBlockContext) GetGcnVgprId(b *SpvBuilder, reg uint32) SpirvId {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnVgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnVgprId(b *SpvBuilder, reg uint32, value SpirvId) {
	b.EmitStore(ctx.GetGcnVgprPtr(b, reg), value)
}

func (ctx *SpirvBlockContext) GetGcnSpecialId(specialId SpirvId) SpirvId {
	id := ctx.GcnSpecialIds[specialId]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn special id %d is zero", specialId))
	}
	id.Used = true
	ctx.GcnSpecialIds[specialId] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnConstId(constId SpirvId) SpirvId {
	id := ctx.GcnConstIds[constId]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn const id %d is zero", constId))
	}
	id.Used = true
	ctx.GcnConstIds[constId] = id

	return id.Id
}

func (ctx *SpirvBlockContext) EmitDebugPrintRegisters(b *SpvBuilder) {
	formatId := b.EmitString("Vertex %d: SGPRs: %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x\n" +
		"VGPRs: %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x\n")
	vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

	args := make([]SpirvId, 0, 33)
	args = append(args, formatId, vertexIndexId)
	for i := uint32(0); i < 16; i++ {
		args = append(args, ctx.GetGcnSgprId(b, i))
	}
	for i := uint32(0); i < 16; i++ {
		args = append(args, ctx.GetGcnVgprId(b, i))
	}

	b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1, args...)
}

// GetRegisterPointer returns the result ID of the pointer to the given register.
func (ctx *SpirvBlockContext) GetRegisterPointer(b *SpvBuilder, op uint32) SpirvId {
	switch {
	case op >= gcnSpec.OpSgpr0 && op <= gcnSpec.OpSgpr103:
		return ctx.GetGcnSgprPtr(b, op)
	case op >= gcnSpec.OpFlatScratchLo && op <= gcnSpec.OpExecHi:
		return ctx.GetGcnSpecialId(SpirvId(op - gcnSpec.OpFlatScratchLo))
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return ctx.GetGcnSpecialId(SpirvId(op-gcnSpec.OpVccz) + GcnSpecIdVccz)
	case op >= gcnSpec.OpVgpr0 && op <= gcnSpec.OpVgpr255:
		return ctx.GetGcnVgprPtr(b, op-gcnSpec.OpVgpr0)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// LoadRegisterPointer loads the value from the given register pointer.
func (ctx *SpirvBlockContext) LoadRegisterPointer(b *SpvBuilder, op uint32) SpirvId {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetRegisterPointer(b, op))
}

// StoreRegisterPointer stores the given value into the given register pointer.
func (ctx *SpirvBlockContext) StoreRegisterPointer(b *SpvBuilder, op uint32, value SpirvId) {
	b.EmitStore(ctx.GetRegisterPointer(b, op), value)
}

func (ctx *SpirvBlockContext) StoreRegisterPointerMasked(b *SpvBuilder, reg uint32, newValue SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeBool := ctx.GetId(BlockContextIdTypeBool)

	// Get current EXEC mask.
	execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
	exec64 := ctx.Pack64(b, execLo, execHi)

	// Get the current thread's bit (bitMask = 1 << SubgroupLocalInvocationId).
	subgroupId := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
	subgroupId64 := b.EmitUConvert(typeUint64, subgroupId)
	subgroupMask := b.EmitShiftLeftLogical(typeUint64, ctx.GetConstId(ConstId64Uint1), subgroupId64)

	// Check if this thread is active (exec64 & bitMask != 0).
	masked := b.EmitBitwiseAnd(typeUint64, exec64, subgroupMask)
	isActive := b.EmitINotEqual(typeBool, masked, ctx.GetConstId(ConstId64Uint0))

	// Load old value from the register.
	oldValue := ctx.LoadRegisterPointer(b, reg)

	// If active, take newValue. If inactive, keep oldValue.
	finalValue := b.EmitSelect(typeUint, isActive, newValue, oldValue)

	ctx.StoreRegisterPointer(b, reg, finalValue)
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	switch {
	case op >= gcnSpec.OpSgpr0 && op <= gcnSpec.OpSgpr103:
		return ctx.GetGcnSgprId(b, op)
	case op >= gcnSpec.OpFlatScratchLo && op <= gcnSpec.OpExecHi:
		return b.EmitLoad(typeUint, ctx.GetGcnSpecialId(SpirvId(op-gcnSpec.OpFlatScratchLo)))
	case op >= gcnSpec.OpInt0 && op <= gcnSpec.OpFloatNeg40:
		return ctx.GetGcnConstId(SpirvId(op - gcnSpec.OpInt0))
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return b.EmitLoad(typeUint, ctx.GetGcnSpecialId(SpirvId(op-gcnSpec.OpVccz)+GcnSpecIdVccz))
	case op == gcnSpec.OpLiteral:
		return b.EmitConstantUint(typeUint, literal)
	case op >= gcnSpec.OpVgpr0 && op <= gcnSpec.OpVgpr255:
		return ctx.GetGcnVgprId(b, op-gcnSpec.OpVgpr0)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperand64Value returns the result IDs of the low and high parts of the value of the given 64-bit operand.
func (ctx *SpirvBlockContext) GetOperand64Value(b *SpvBuilder, op uint32, literal uint32) (SpirvId, SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	switch {
	case op >= gcnSpec.OpSgpr0 && op <= gcnSpec.OpSgpr103:
		return ctx.GetGcnSgprId(b, op), ctx.GetGcnSgprId(b, op+1)
	case op >= gcnSpec.OpFlatScratchLo && op <= gcnSpec.OpExecHi:
		return b.EmitLoad(typeUint, ctx.GetGcnSpecialId(SpirvId(op-gcnSpec.OpFlatScratchLo))), b.EmitLoad(typeUint, ctx.GetGcnSpecialId(SpirvId(op+1-gcnSpec.OpFlatScratchLo)))
	case op >= gcnSpec.OpVgpr0 && op <= gcnSpec.OpVgpr255:
		return ctx.GetGcnVgprId(b, op-gcnSpec.OpVgpr0), ctx.GetGcnVgprId(b, op-gcnSpec.OpVgpr0+1)
	case op >= gcnSpec.OpInt0 && op <= gcnSpec.OpPosInt64:
		return ctx.GetGcnConstId(SpirvId(op - gcnSpec.OpInt0)), ctx.GetConstId(ConstIdUint0)
	case op >= gcnSpec.OpNegInt1 && op <= gcnSpec.OpNegInt16:
		return ctx.GetGcnConstId(SpirvId(op - gcnSpec.OpInt0)), ctx.GetConstId(ConstIdUintFFFFFFFF)
	case op >= gcnSpec.OpFloat05 && op <= gcnSpec.OpFloatNeg40:
		return ctx.GetGcnConstId(SpirvId(op - gcnSpec.OpInt0)), ctx.GetConstId(ConstIdUint0)
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return b.EmitLoad(typeUint, ctx.GetGcnSpecialId(SpirvId(op-gcnSpec.OpVccz)+GcnSpecIdVccz)), ctx.GetConstId(ConstIdUint0)
	case op == gcnSpec.OpLiteral:
		return b.EmitConstantUint(typeUint, literal), ctx.GetConstId(ConstIdUint0)
	}

	panic(fmt.Sprintf("unknown 64-bit op %d", op))
}

// GetOperandUintValue returns the result ID of the value of the given operand as a uint.
func (ctx *SpirvBlockContext) GetOperandUintValue(b *SpvBuilder, op uint32, literal uint32) SpirvId {
	return ctx.GetOperandValue(b, op, literal)
}

// GetOperandIntValue returns the result ID of the value of the given operand as an int.
func (ctx *SpirvBlockContext) GetOperandIntValue(b *SpvBuilder, op uint32, literal uint32) SpirvId {
	return b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), ctx.GetOperandValue(b, op, literal))
}

// GetOperandFloatValue returns the result ID of the value of the given operand as a float.
func (ctx *SpirvBlockContext) GetOperandFloatValue(b *SpvBuilder, op uint32, literal uint32) SpirvId {
	return b.EmitBitcast(ctx.GetId(BlockContextIdTypeFloat), ctx.GetOperandValue(b, op, literal))
}

// TestMask returns a boolean result ID of (val & mask) != 0.
func (ctx *SpirvBlockContext) TestMask(b *SpvBuilder, value SpirvId, mask uint32) SpirvId {
	maskId := b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), mask)
	andId := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), value, maskId)
	return b.EmitINotEqual(ctx.GetId(BlockContextIdTypeBool), andId, ctx.GetConstId(ConstIdUint0))
}

// Pack64 combines two 32-bit values into one 64-bit value.
func (ctx *SpirvBlockContext) Pack64(b *SpvBuilder, lo, hi SpirvId) SpirvId {
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)

	lo64 := b.EmitUConvert(typeUint64, lo)
	hi64 := b.EmitUConvert(typeUint64, hi)
	hiShifted := b.EmitShiftLeftLogical(typeUint64, hi64, ctx.GetConstId(ConstId64Uint32))
	return b.EmitBitwiseOr(typeUint64, lo64, hiShifted)
}

// LoadPushConstantValue loads a value from the push constant at the given index.
func (ctx *SpirvBlockContext) LoadPushConstantValue(b *SpvBuilder, i uint32) SpirvId {
	var valType, ptrType SpirvId
	switch i {
	case PushConstantUserDataAddress:
		valType = ctx.GetId(BlockContextIdPtrPsbUint)
		ptrType = ctx.GetId(BlockContextIdPtrPcPsbUint)
	case PushConstantOnionMemoryBaseAddress, PushConstantGarlicMemoryBaseAddress:
		valType = ctx.GetId(BlockContextIdTypeUint64)
		ptrType = ctx.GetId(BlockContextIdPtrPcUint64)
	case PushConstantUserSgprCount, PushConstantShaderRsrc2, PushConstantVteControl, PushConstantClipControl:
		valType = ctx.GetId(BlockContextIdTypeUint)
		ptrType = ctx.GetId(BlockContextIdPtrPcUint)
	case PushConstantGbHorzClipAdj, PushConstantGbVertClipAdj:
		valType = ctx.GetId(BlockContextIdTypeFloat)
		ptrType = ctx.GetId(BlockContextIdPtrPcFloat)
	default:
		panic(fmt.Sprintf("unknown push constant index %d", i))
	}

	ptr := b.EmitAccessChain(ptrType, ctx.GetId(BlockContextIdPcVar), b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), i))
	return b.EmitLoad(valType, ptr)
}

// LoadPsInputParameter loads a pixel shader input parameter.
func (ctx *SpirvBlockContext) LoadPsInputParameter(b *SpvBuilder, i uint32) SpirvId {
	typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
	control := ctx.PsInputControls[i]
	offset := control & 0x3F
	if match := offset&0x20 == 0; match {
		ptr := ctx.GetId(BlockContextIdParamIn0 + SpirvId(i))
		return b.EmitLoad(typeV4Float, ptr)
	}

	// No vertex shader match, use default value.
	idZeroF := ctx.GetConstId(ConstIdFloat0)
	idOneF := ctx.GetConstId(ConstIdFloat1)
	defaultValue := (control >> 8) & 3
	switch defaultValue {
	case 0: // 0.0, 0.0, 0.0, 0.0
		return b.EmitConstantComposite(typeV4Float, idZeroF, idZeroF, idZeroF, idZeroF)
	case 1: // 0.0, 0.0, 0.0, 1.0
		return b.EmitConstantComposite(typeV4Float, idZeroF, idZeroF, idZeroF, idOneF)
	case 2: // 1.0, 1.0, 1.0, 0.0
		return b.EmitConstantComposite(typeV4Float, idOneF, idOneF, idOneF, idZeroF)
	case 3: // 1.0, 1.0, 1.0, 1.0
		return b.EmitConstantComposite(typeV4Float, idOneF, idOneF, idOneF, idOneF)
	}

	return ctx.GetId(BlockContextIdZeroVec4)
}

// TranslateAddress translates a PS4 address into a memory buffer address.
func (ctx *SpirvBlockContext) TranslateAddress(b *SpvBuilder, address SpirvId) SpirvId {
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)

	// Clean address using 40-bit mask.
	mask := b.EmitConstantUint64(typeUint64, 0xFFFFFFFFFF)
	cleanAddress := b.EmitBitwiseAnd(typeUint64, address, mask)

	// Result variable initialization.
	idZero64 := ctx.GetConstId(ConstId64Uint0)
	idPtrFnUint64 := b.EmitTypePointer(spec.SpvStorageFunction, typeUint64)
	idResultVar := b.AllocId()
	b.EmitDeferredLocalVariable(idPtrFnUint64, idResultVar)
	b.EmitStore(idResultVar, idZero64)
	b.EmitName(idResultVar, "translated_address")

	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	idIndexVar := b.AllocId()
	b.EmitDeferredLocalVariable(idPtrFnUint, idIndexVar)
	b.EmitStore(idIndexVar, ctx.GetConstId(ConstIdUint0))
	b.EmitName(idIndexVar, "translation_loop_index")

	// Labels for the loop.
	loopHead := b.AllocId()
	loopBody := b.AllocId()
	loopContinue := b.AllocId()
	loopMerge := b.AllocId()

	b.EmitBranch(loopHead)
	b.EmitLabel(loopHead)
	b.EmitLoopMerge(loopMerge, loopContinue, spec.SpvLoopControlNone)
	b.EmitBranch(loopBody)

	// Loop body.
	b.EmitLabel(loopBody)
	idIndex := b.EmitLoad(typeUint, idIndexVar)

	// Pointers.
	idPtrStorageUint64 := b.EmitTypePointer(spec.SpvStorageStorageBuffer, typeUint64)
	idBufferVar := ctx.GetId(BlockContextIdAddressTranslationBuffer)

	// Entry pointers (0=GuestBase, 1=GuestEnd, 2=DeviceAddress).
	idBasePtr := b.EmitAccessChain(idPtrStorageUint64, idBufferVar, ctx.GetConstId(ConstIdUint0), idIndex, ctx.GetConstId(ConstIdUint0))
	idEndPtr := b.EmitAccessChain(idPtrStorageUint64, idBufferVar, ctx.GetConstId(ConstIdUint0), idIndex, ctx.GetConstId(ConstIdUint1))
	idBdaPtr := b.EmitAccessChain(idPtrStorageUint64, idBufferVar, ctx.GetConstId(ConstIdUint0), idIndex, ctx.GetConstId(ConstIdUint2))

	idBase := b.EmitLoad(typeUint64, idBasePtr)
	idSentinel := b.EmitConstantUint64(typeUint64, ^uint64(0))

	// Check if sentinel reached.
	isSentinel := b.EmitIEqual(typeBool, idBase, idSentinel)
	isSentinelTrue := b.AllocId()
	isSentinelFalse := b.AllocId()
	b.EmitSelectionMerge(isSentinelFalse, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(isSentinel, isSentinelTrue, isSentinelFalse)

	b.EmitLabel(isSentinelTrue)
	b.EmitBranch(loopMerge)

	b.EmitLabel(isSentinelFalse)

	// Check if address is in range.
	idEnd := b.EmitLoad(typeUint64, idEndPtr)
	isGTEBase := b.EmitUGreaterThanEqual(typeBool, cleanAddress, idBase)
	isLTEnd := b.EmitULessThan(typeBool, cleanAddress, idEnd)
	inRange := b.EmitLogicalAnd(typeBool, isGTEBase, isLTEnd)

	inRangeTrue := b.AllocId()
	inRangeFalse := b.AllocId()
	b.EmitSelectionMerge(inRangeFalse, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(inRange, inRangeTrue, inRangeFalse)

	b.EmitLabel(inRangeTrue)
	idBda := b.EmitLoad(typeUint64, idBdaPtr)
	offset := b.EmitISub(typeUint64, cleanAddress, idBase)
	translated := b.EmitIAdd(typeUint64, idBda, offset)
	b.EmitStore(idResultVar, translated)
	b.EmitBranch(loopMerge)

	b.EmitLabel(inRangeFalse)
	b.EmitBranch(loopContinue)

	// Loop continue.
	b.EmitLabel(loopContinue)
	idNextIndex := b.EmitIAdd(typeUint, idIndex, ctx.GetConstId(ConstIdUint1))
	b.EmitStore(idIndexVar, idNextIndex)
	b.EmitBranch(loopHead)

	// Loop merge.
	b.EmitLabel(loopMerge)
	return b.EmitLoad(typeUint64, idResultVar)
}

// EmitDebugPrintfLane emits a debug printf only for a specific subgroup lane.
func (ctx *SpirvBlockContext) EmitDebugPrintfLane(b *SpvBuilder, lane uint32, format string, args ...SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idLane := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
	isLane := b.EmitIEqual(typeBool, idLane, ctx.GetConstId(ConstIdUint0+SpirvId(lane)))

	thenLabel := b.AllocId()
	mergeLabel := b.AllocId()
	b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(isLane, thenLabel, mergeLabel)

	b.EmitLabel(thenLabel)
	ctx.EmitDebugPrintf(b, format, args...)
	b.EmitBranch(mergeLabel)

	b.EmitLabel(mergeLabel)
}

// EmitDebugPrintfLane emits a debug printf.
func (ctx *SpirvBlockContext) EmitDebugPrintf(b *SpvBuilder, format string, args ...SpirvId) {
	ins := make([]SpirvId, 0, len(args)+1)
	ins = append(ins, b.EmitString(format))
	ins = append(ins, args...)
	b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1, ins...)
}
