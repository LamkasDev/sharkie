package spirv

// SpvBuilder accumulates SPIR-V words so we can assemble them in the correct order.
type SpvBuilder struct {
	nextId            uint32
	caps              []uint32 // OpCapability
	exts              []uint32 // OpExtension / OpExtInstImport
	memModel          []uint32 // OpMemoryModel
	entryPts          []uint32 // OpEntryPoint
	execModes         []uint32 // OpExecutionMode
	debugStrings      []uint32 // OpString / OpSource / ...
	debugNames        []uint32 // OpName / OpMemberName
	annots            []uint32 // OpDecorate / OpMemberDecorate
	types             []uint32 // types, constants, global variables
	deferredConstants []uint32 // deferred constants
	deferredLocalVars []uint32 // deferred local variables
	code              []uint32 // function bodies
}

// NewSpvBuilder creates a new SpvBuilder.
func NewSpvBuilder() *SpvBuilder {
	return &SpvBuilder{
		nextId: 1,
	}
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

// EmitExtension emits OpExtension.
func (b *SpvBuilder) EmitExtension(name string) {
	b.instr(&b.exts, SpvOpExtension, spirvString(name)...)
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

// EmitName emits OpName.
func (b *SpvBuilder) EmitName(target uint32, name string) {
	operands := append([]uint32{target}, spirvString(name)...)
	b.instr(&b.debugNames, SpvOpName, operands...)
}

// EmitString emits OpString and returns the result ID.
func (b *SpvBuilder) EmitString(s string) uint32 {
	id := b.AllocId()
	operands := append([]uint32{id}, spirvString(s)...)
	b.instr(&b.code, SpvOpString, operands...)
	return id
}

// EmitLine emits OpLine.
func (b *SpvBuilder) EmitLine(fileID, line, column uint32) {
	b.instr(&b.code, SpvOpLine, fileID, line, column)
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

// EmitReturn emits OpReturn.
func (b *SpvBuilder) EmitReturn() {
	b.instr(&b.code, SpvOpReturn)
}

// EmitAccessChain emits OpAccessChain and returns the result pointer ID.
func (b *SpvBuilder) EmitAccessChain(resultType, base uint32, indices ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, base}, indices...)
	b.instr(&b.code, SpvOpAccessChain, ops...)
	return id
}

// EmitPtrAccessChain emits OpPtrAccessChain and returns the result pointer ID.
func (b *SpvBuilder) EmitPtrAccessChain(resultType, base, element uint32, indices ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, base, element}, indices...)
	b.instr(&b.code, SpvOpPtrAccessChain, ops...)
	return id
}

// EmitUConvert emits OpUConvert and returns the result ID.
func (b *SpvBuilder) EmitUConvert(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpUConvert, resultType, id, operand)
	return id
}

// EmitBitcast emits OpBitcast and returns the result ID.
func (b *SpvBuilder) EmitBitcast(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpBitcast, resultType, id, operand)
	return id
}

// EmitConvertUToF emits EmitConvertUToF and returns the result ID.
func (b *SpvBuilder) EmitConvertUToF(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpConvertUToF, resultType, id, operand)
	return id
}

// EmitConvertUToPtr emits OpConvertUToPtr and returns the result ID.
func (b *SpvBuilder) EmitConvertUToPtr(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpConvertUToPtr, resultType, id, operand)
	return id
}

// EmitIsNan emits OpIsNan and returns the result ID.
func (b *SpvBuilder) EmitIsNan(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpIsNan, resultType, id, operand)
	return id
}

// EmitGroupNonUniformBallot emits OpGroupNonUniformBallot and returns the result ID.
func (b *SpvBuilder) EmitGroupNonUniformBallot(resultType, scope, predicate uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpGroupNonUniformBallot, resultType, id, scope, predicate)
	return id
}

// EmitImageFetch emits OpImageFetch and returns the result ID.
func (b *SpvBuilder) EmitImageFetch(resultType, image, coordinate uint32, imageOperands ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, image, coordinate}, imageOperands...)
	b.instr(&b.code, SpvOpImageFetch, ops...)
	return id
}

// EmitSelect emits OpSelect and returns the result ID.
func (b *SpvBuilder) EmitSelect(resultType, condition, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpSelect, resultType, id, condition, op1, op2)
	return id
}

// EmitExtInst emits OpExtInst (like pack instructions) and returns the result ID.
func (b *SpvBuilder) EmitExtInst(resultType, setID, instruction uint32, operands ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, setID, instruction}, operands...)
	b.instr(&b.code, SpvOpExtInst, ops...)
	return id
}

// EmitExtInstImport emits OpExtInstImport and returns the result ID.
func (b *SpvBuilder) EmitExtInstImport(name string) uint32 {
	id := b.AllocId()
	b.instr(&b.exts, SpvOpExtInstImport, append([]uint32{id}, spirvString(name)...)...)
	return id
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
	out = append(out, b.debugStrings...)
	out = append(out, b.debugNames...)
	out = append(out, b.annots...)
	out = append(out, b.types...)
	out = append(out, b.deferredConstants...)

	// Insert local variables after the first label in the code section.
	code := b.code
	for i := 0; i < len(code); {
		wordCount := code[i] >> 16
		opCode := code[i] & 0xFFFF
		if opCode == SpvOpLabel {
			// Found the first label, insert local variables right after it.
			i += int(wordCount)
			newCode := make([]uint32, 0, len(code)+len(b.deferredLocalVars))
			newCode = append(newCode, code[:i]...)
			newCode = append(newCode, b.deferredLocalVars...)
			newCode = append(newCode, code[i:]...)
			code = newCode
			break
		}
		i += int(wordCount)
	}

	out = append(out, code...)

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
