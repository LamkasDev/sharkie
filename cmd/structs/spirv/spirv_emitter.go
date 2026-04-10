package spirv

import (
	"math"

	"go101.org/nstd"
)

// SpvBuilder accumulates SPIR-V words so we can assemble them in the correct order.
type SpvBuilder struct {
	nextId    uint32
	caps      []uint32 // OpCapability
	exts      []uint32 // OpExtension / OpExtInstImport
	memModel  []uint32 // OpMemoryModel
	entryPts  []uint32 // OpEntryPoint
	execModes []uint32 // OpExecutionMode
	annots    []uint32 // OpDecorate / OpMemberDecorate
	types     []uint32 // types, constants, global variables
	code      []uint32 // function bodies
}

// NewSpvBuilder creates a new SpvBuilder.
func NewSpvBuilder() *SpvBuilder {
	return &SpvBuilder{nextId: 1}
}

// AllocId returns the next available SPIR-V ID.
func (b *SpvBuilder) AllocId() uint32 {
	id := b.nextId
	b.nextId++

	return id
}

// instr appends one SPIR-V instruction to section.
func (b *SpvBuilder) instr(section *[]uint32, opcode uint32, operands ...uint32) {
	wc := uint32(1 + len(operands))
	*section = append(*section, (wc<<16)|opcode)
	*section = append(*section, operands...)
}

// EmitCapability emits OpCapability.
func (b *SpvBuilder) EmitCapability(cap uint32) {
	b.instr(&b.caps, SpvOpCapability, cap)
}

// EmitMemoryModel emits OpMemoryModel.
func (b *SpvBuilder) EmitMemoryModel(addrModel, memModel uint32) {
	b.instr(&b.memModel, SpvOpMemoryModel, addrModel, memModel)
}

// EmitEntryPoint emits OpEntryPoint (optional input/output variable IDs).
func (b *SpvBuilder) EmitEntryPoint(execModel, funcID uint32, name string, interfaceIDs ...uint32) {
	operands := []uint32{execModel, funcID}
	operands = append(operands, spirvString(name)...)
	operands = append(operands, interfaceIDs...)
	b.instr(&b.entryPts, SpvOpEntryPoint, operands...)
}

// EmitExecutionMode emits OpExecutionMode.
func (b *SpvBuilder) EmitExecutionMode(funcID, mode uint32, args ...uint32) {
	operands := append([]uint32{funcID, mode}, args...)
	b.instr(&b.execModes, SpvOpExecutionMode, operands...)
}

// EmitDecorate decorates a target type (optional extra operands).
func (b *SpvBuilder) EmitDecorate(target, decoration uint32, values ...uint32) {
	operands := append([]uint32{target, decoration}, values...)
	b.instr(&b.annots, SpvOpDecorate, operands...)
}

// EmitDecorate decorates a target struct member (optional extra operands).
func (b *SpvBuilder) EmitMemberDecorate(structType, member, decoration uint32, values ...uint32) {
	operands := append([]uint32{structType, member, decoration}, values...)
	b.instr(&b.annots, SpvOpMemberDecorate, operands...)
}

// EmitTypeInt declares a void type.
func (b *SpvBuilder) EmitTypeVoid() uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeVoid, id)
	return id
}

// EmitTypeInt declares a boolean type.
func (b *SpvBuilder) EmitTypeBool() uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeBool, id)
	return id
}

// EmitTypeInt declares an integer type.
func (b *SpvBuilder) EmitTypeInt(width uint32, signed bool) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeInt, id, width, uint32(nstd.Btoi(signed)))
	return id
}

// EmitTypeFloat declares a float type.
func (b *SpvBuilder) EmitTypeFloat(width uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeFloat, id, width)
	return id
}

// EmitTypeVector declares a vector elementType[count].
func (b *SpvBuilder) EmitTypeVector(elementType, count uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeVector, id, elementType, count)
	return id
}

// EmitTypeArray declares an array elementType[length] (length is the ID of an integer constant).
func (b *SpvBuilder) EmitTypeArray(elementType, lengthID uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeArray, id, elementType, lengthID)
	return id
}

// EmitTypeStruct declares a struct with given member types.
func (b *SpvBuilder) EmitTypeStruct(memberTypes ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{id}, memberTypes...)
	b.instr(&b.types, SpvOpTypeStruct, operands...)
	return id
}

// EmitTypePointer declares a pointer to pointerType in the given storage class.
func (b *SpvBuilder) EmitTypePointer(storageClass, pointerType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypePointer, id, storageClass, pointerType)
	return id
}

// EmitTypeFunction emits OpTypeFunction.
func (b *SpvBuilder) EmitTypeFunction(returnType uint32, paramTypes ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{id, returnType}, paramTypes...)
	b.instr(&b.types, SpvOpTypeFunction, operands...)
	return id
}

// EmitConstantUint emits OpConstant for a 32-bit unsigned integer.
func (b *SpvBuilder) EmitConstantUint(uintType, value uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, uintType, id, value)
	return id
}

// EmitConstantFalse emits OpConstantFalse.
func (b *SpvBuilder) EmitConstantFalse(boolType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstantFalse, boolType, id)
	return id
}

// EmitConstantFloat emits OpConstant for a float.
func (b *SpvBuilder) EmitConstantFloat(floatType uint32, value float32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, floatType, id, math.Float32bits(value))
	return id
}

// EmitConstantComposite emits OpConstantComposite for a vector or composite constant.
func (b *SpvBuilder) EmitConstantComposite(resultType uint32, constituents ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{resultType, id}, constituents...)
	b.instr(&b.types, SpvOpConstantComposite, operands...)
	return id
}

// EmitVariable emits a global OpVariable.
func (b *SpvBuilder) EmitVariable(ptrType, storageClass uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpVariable, ptrType, id, storageClass)
	return id
}

// EmitFunction emits OpFunction.
func (b *SpvBuilder) EmitFunction(returnType, funcControl, funcType, funcID uint32) {
	b.instr(&b.code, SpvOpFunction, returnType, funcID, funcControl, funcType)
}

// EmitFunctionEnd emits OpFunctionEnd.
func (b *SpvBuilder) EmitFunctionEnd() {
	b.instr(&b.code, SpvOpFunctionEnd)
}

// EmitLabel emits OpLabel.
func (b *SpvBuilder) EmitLabel(id uint32) {
	b.instr(&b.code, SpvOpLabel, id)
}

// EmitStore emits OpStore.
func (b *SpvBuilder) EmitStore(pointer, object uint32) {
	b.instr(&b.code, SpvOpStore, pointer, object)
}

// EmitBranch emits OpBranch.
func (b *SpvBuilder) EmitBranch(targetLabel uint32) {
	b.instr(&b.code, SpvOpBranch, targetLabel)
}

// EmitBranchConditional emits OpBranchConditional.
func (b *SpvBuilder) EmitBranchConditional(condID, trueLabel, falseLabel uint32) {
	b.instr(&b.code, SpvOpBranchConditional, condID, trueLabel, falseLabel)
}

// EmitSelectionMerge emits OpSelectionMerge (must appear immediately before the OpBranchConditional or OpSwitch it governs).
func (b *SpvBuilder) EmitSelectionMerge(mergeBlock, selectionControl uint32) {
	b.instr(&b.code, SpvOpSelectionMerge, mergeBlock, selectionControl)
}

// EmitLoopMerge emits OpLoopMerge (must appear immediately before the branch instruction that closes the loop header).
func (b *SpvBuilder) EmitLoopMerge(mergeBlock, continueBlock, loopControl uint32) {
	b.instr(&b.code, SpvOpLoopMerge, mergeBlock, continueBlock, loopControl)
}

// EmitReturn emits OpReturn.
func (b *SpvBuilder) EmitReturn() {
	b.instr(&b.code, SpvOpReturn)
}

// EmitUnreachable emits OpUnreachable.
func (b *SpvBuilder) EmitUnreachable() {
	b.instr(&b.code, SpvOpUnreachable)
}

// Assemble combines all sections in SPIR-V specification order and returns the complete module as []uint32 ready for vkCreateShaderModule.
func (b *SpvBuilder) Assemble() []uint32 {
	var out []uint32
	out = append(out, SpvMagic, SpvVersion, SpvGen, b.nextId, 0)
	out = append(out, b.caps...)
	out = append(out, b.exts...)
	out = append(out, b.memModel...)
	out = append(out, b.entryPts...)
	out = append(out, b.execModes...)
	out = append(out, b.annots...)
	out = append(out, b.types...)
	out = append(out, b.code...)

	return out
}

// spirvString encodes a Go string as SPIR-V words (null-terminated and zero-padded to the next 4-byte boundary).
func spirvString(s string) []uint32 {
	// Append null terminator then pad to multiple of 4.
	b := []byte(s)
	b = append(b, 0)
	for len(b)%4 != 0 {
		b = append(b, 0)
	}

	words := make([]uint32, len(b)/4)
	for i := range words {
		words[i] = uint32(b[i*4]) |
			uint32(b[i*4+1])<<8 |
			uint32(b[i*4+2])<<16 |
			uint32(b[i*4+3])<<24
	}

	return words
}

// SpvWordsToBytes converts a []uint32 SPIR-V module to []byte slice.
func SpvWordsToBytes(words []uint32) []byte {
	out := make([]byte, len(words)*4)
	for i, w := range words {
		out[i*4+0] = byte(w)
		out[i*4+1] = byte(w >> 8)
		out[i*4+2] = byte(w >> 16)
		out[i*4+3] = byte(w >> 24)
	}

	return out
}
