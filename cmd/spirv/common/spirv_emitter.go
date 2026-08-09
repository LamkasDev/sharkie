package common

import "github.com/LamkasDev/sharkie/cmd/spirv/spec"

// SpvBuilder accumulates SPIR-V words so we can assemble them in the correct order.
type SpvBuilder struct {
	nextId            SpirvId
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
func (b *SpvBuilder) AllocId() SpirvId {
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
	b.instr(&b.caps, spec.SpvOpCapability, cap)
}

// EmitExtension emits OpExtension.
func (b *SpvBuilder) EmitExtension(name string) {
	b.instr(&b.exts, spec.SpvOpExtension, spirvString(name)...)
}

// EmitMemoryModel emits OpMemoryModel.
func (b *SpvBuilder) EmitMemoryModel(addrModel, memModel uint32) {
	b.instr(&b.memModel, spec.SpvOpMemoryModel, addrModel, memModel)
}

// EmitEntryPoint emits OpEntryPoint (optional input/output variable IDs).
func (b *SpvBuilder) EmitEntryPoint(execModel uint32, funcId SpirvId, name string, interfaceIds ...SpirvId) {
	operands := []uint32{execModel, uint32(funcId)}
	operands = append(operands, spirvString(name)...)
	for _, id := range interfaceIds {
		operands = append(operands, uint32(id))
	}
	b.instr(&b.entryPts, spec.SpvOpEntryPoint, operands...)
}

// EmitExecutionMode emits OpExecutionMode.
func (b *SpvBuilder) EmitExecutionMode(funcId SpirvId, mode uint32, args ...uint32) {
	operands := append([]uint32{uint32(funcId), mode}, args...)
	b.instr(&b.execModes, spec.SpvOpExecutionMode, operands...)
}

// EmitName emits OpName.
func (b *SpvBuilder) EmitName(target SpirvId, name string) {
	operands := append([]uint32{uint32(target)}, spirvString(name)...)
	b.instr(&b.debugNames, spec.SpvOpName, operands...)
}

// EmitString emits OpString and returns the result ID.
func (b *SpvBuilder) EmitString(s string) SpirvId {
	id := b.AllocId()
	operands := append([]uint32{uint32(id)}, spirvString(s)...)
	b.instr(&b.debugStrings, spec.SpvOpString, operands...)
	return id
}

// EmitLine emits OpLine.
func (b *SpvBuilder) EmitLine(fileId SpirvId, line, column uint32) {
	b.instr(&b.code, spec.SpvOpLine, uint32(fileId), line, column)
}

// EmitDecorate decorates a target type (optional extra operands).
func (b *SpvBuilder) EmitDecorate(target SpirvId, decoration uint32, values ...uint32) {
	operands := append([]uint32{uint32(target), decoration}, values...)
	b.instr(&b.annots, spec.SpvOpDecorate, operands...)
}

// EmitDecorate decorates a target struct member (optional extra operands).
func (b *SpvBuilder) EmitMemberDecorate(structType SpirvId, member, decoration uint32, values ...uint32) {
	operands := append([]uint32{uint32(structType), member, decoration}, values...)
	b.instr(&b.annots, spec.SpvOpMemberDecorate, operands...)
}

// EmitFunction emits OpFunction.
func (b *SpvBuilder) EmitFunction(returnType, funcId SpirvId, funcControl uint32, funcType SpirvId) {
	b.instr(&b.code, spec.SpvOpFunction, uint32(returnType), uint32(funcId), funcControl, uint32(funcType))
}

// EmitFunctionEnd emits OpFunctionEnd.
func (b *SpvBuilder) EmitFunctionEnd() {
	b.instr(&b.code, spec.SpvOpFunctionEnd)
}

// EmitLabel emits OpLabel.
func (b *SpvBuilder) EmitLabel(id SpirvId) {
	b.instr(&b.code, spec.SpvOpLabel, uint32(id))
}

// EmitReturn emits OpReturn.
func (b *SpvBuilder) EmitReturn() {
	b.instr(&b.code, spec.SpvOpReturn)
}

// EmitKill emits OpKill.
func (b *SpvBuilder) EmitKill() {
	b.instr(&b.code, spec.SpvOpKill)
}

// EmitAccessChain emits OpAccessChain and returns the result pointer ID.
func (b *SpvBuilder) EmitAccessChain(resultType, base SpirvId, indices ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := append([]uint32{uint32(resultType), uint32(id), uint32(base)})
	for _, ind := range indices {
		operands = append(operands, uint32(ind))
	}
	b.instr(&b.code, spec.SpvOpAccessChain, operands...)
	return id
}

// EmitPtrAccessChain emits OpPtrAccessChain and returns the result pointer ID.
func (b *SpvBuilder) EmitPtrAccessChain(resultType, base, element SpirvId, indices ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := append([]uint32{uint32(resultType), uint32(id), uint32(base), uint32(element)})
	for _, ind := range indices {
		operands = append(operands, uint32(ind))
	}
	b.instr(&b.code, spec.SpvOpPtrAccessChain, operands...)
	return id
}

// EmitUConvert emits OpUConvert and returns the result ID.
func (b *SpvBuilder) EmitUConvert(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpUConvert, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitBitcast emits OpBitcast and returns the result ID.
func (b *SpvBuilder) EmitBitcast(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpBitcast, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitConvertUToF emits OpConvertUToF and returns the result ID.
func (b *SpvBuilder) EmitConvertUToF(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpConvertUToF, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitConvertFToU emits OpConvertFToU and returns the result ID.
func (b *SpvBuilder) EmitConvertFToU(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpConvertFToU, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitConvertFToS emits OpConvertFToS and returns the result ID.
func (b *SpvBuilder) EmitConvertFToS(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpConvertFToS, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitConvertSToF emits OpConvertSToF and returns the result ID.
func (b *SpvBuilder) EmitConvertSToF(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpConvertSToF, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitConvertUToPtr emits OpConvertUToPtr and returns the result ID.
func (b *SpvBuilder) EmitConvertUToPtr(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpConvertUToPtr, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitIsNan emits OpIsNan and returns the result ID.
func (b *SpvBuilder) EmitIsNan(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpIsNan, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitGroupNonUniformBallot emits OpGroupNonUniformBallot and returns the result ID.
func (b *SpvBuilder) EmitGroupNonUniformBallot(resultType, scope, predicate SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpGroupNonUniformBallot, uint32(resultType), uint32(id), uint32(scope), uint32(predicate))
	return id
}

// EmitImageFetch emits OpImageFetch and returns the result ID.
func (b *SpvBuilder) EmitImageFetch(resultType, image, coordinate SpirvId, mask uint32, imageOperands ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := []uint32{uint32(resultType), uint32(id), uint32(image), uint32(coordinate)}
	if mask != 0 {
		operands = append(operands, mask)
		for _, op := range imageOperands {
			operands = append(operands, uint32(op))
		}
	}
	b.instr(&b.code, spec.SpvOpImageFetch, operands...)
	return id
}

// EmitImageRead emits OpImageRead and returns the result ID.
func (b *SpvBuilder) EmitImageRead(resultType, image, coordinate SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageRead, uint32(resultType), uint32(id), uint32(image), uint32(coordinate))
	return id
}

// EmitImage emits OpImage and returns the result ID.
func (b *SpvBuilder) EmitImage(resultType, sampledImage SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImage, uint32(resultType), uint32(id), uint32(sampledImage))
	return id
}

// EmitSampledImage emits OpSampledImage and returns the result ID.
func (b *SpvBuilder) EmitSampledImage(resultType, image, sampler SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpSampledImage, uint32(resultType), uint32(id), uint32(image), uint32(sampler))
	return id
}

// EmitImageSampleImplicitLod emits OpImageSampleImplicitLod without any optional image operands.
func (b *SpvBuilder) EmitImageSampleImplicitLod(resultType, sampledImage, coordinate SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageSampleImplicitLod, uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate))
	return id
}

// EmitImageSampleImplicitLodOperands emits OpImageSampleImplicitLod with a specific image operand mask and variadic arguments.
func (b *SpvBuilder) EmitImageSampleImplicitLodOperands(resultType, sampledImage, coordinate SpirvId, imageOperandsMask uint32, operands ...SpirvId) SpirvId {
	id := b.AllocId()
	args := []uint32{uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate), imageOperandsMask}
	for _, op := range operands {
		args = append(args, uint32(op))
	}
	b.instr(&b.code, spec.SpvOpImageSampleImplicitLod, args...)
	return id
}

// EmitImageSampleExplicitLod emits OpImageSampleExplicitLod with the standard LOD mask and operand.
func (b *SpvBuilder) EmitImageSampleExplicitLod(resultType, sampledImage, coordinate, lod SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageSampleExplicitLod, uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate), uint32(spec.SpvImageOperandsLodMask), uint32(lod))
	return id
}

// EmitImageSampleExplicitLodOperands emits OpImageSampleExplicitLod with a custom image operand mask and variadic arguments.
func (b *SpvBuilder) EmitImageSampleExplicitLodOperands(resultType, sampledImage, coordinate SpirvId, imageOperandsMask uint32, operands ...SpirvId) SpirvId {
	id := b.AllocId()
	args := []uint32{uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate), imageOperandsMask}
	for _, op := range operands {
		args = append(args, uint32(op))
	}
	b.instr(&b.code, spec.SpvOpImageSampleExplicitLod, args...)
	return id
}

// EmitImageQuerySizeLod emits OpImageQuerySizeLod and returns the result ID.
func (b *SpvBuilder) EmitImageQuerySizeLod(resultType, image, lod SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageQuerySizeLod, uint32(resultType), uint32(id), uint32(image), uint32(lod))
	return id
}

// EmitImageQueryLod emits OpImageQueryLod and returns the result id (a vec2 float).
func (b *SpvBuilder) EmitImageQueryLod(resultType, sampledImage, coordinate SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageQueryLod, uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate))
	return id
}

// EmitImageWrite emits OpImageWrite.
func (b *SpvBuilder) EmitImageWrite(image, coordinate, texel SpirvId) {
	b.instr(&b.code, spec.SpvOpImageWrite, uint32(image), uint32(coordinate), uint32(texel))
}

// EmitAtomicLoad emits OpAtomicLoad and returns the result ID.
func (b *SpvBuilder) EmitAtomicLoad(resultType, pointer, scope, semantics SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpAtomicLoad, uint32(resultType), uint32(id), uint32(pointer), uint32(scope), uint32(semantics))
	return id
}

// EmitAtomicStore emits OpAtomicStore.
func (b *SpvBuilder) EmitAtomicStore(pointer, scope, semantics, value SpirvId) {
	b.instr(&b.code, spec.SpvOpAtomicStore, uint32(pointer), uint32(scope), uint32(semantics), uint32(value))
}

// EmitAtomicExchange emits OpAtomicExchange and returns the result ID.
func (b *SpvBuilder) EmitAtomicExchange(resultType, pointer, scope, semantics, value SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpAtomicExchange, uint32(resultType), uint32(id), uint32(pointer), uint32(scope), uint32(semantics), uint32(value))
	return id
}

// EmitAtomicIAdd emits OpAtomicIAdd and returns the result ID.
func (b *SpvBuilder) EmitAtomicIAdd(resultType, pointer, scope, semantics, value SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpAtomicIAdd, uint32(resultType), uint32(id), uint32(pointer), uint32(scope), uint32(semantics), uint32(value))
	return id
}

// EmitAtomicCompareExchange emits OpAtomicCompareExchange and returns the result ID.
func (b *SpvBuilder) EmitAtomicCompareExchange(resultType, pointer, scope, semanticsEqual, semanticsUnequal, value, comparator SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpAtomicCompareExchange, uint32(resultType), uint32(id), uint32(pointer), uint32(scope), uint32(semanticsEqual), uint32(semanticsUnequal), uint32(value), uint32(comparator))
	return id
}

// EmitSelect emits OpSelect and returns the result ID.
func (b *SpvBuilder) EmitSelect(resultType, condition, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpSelect, uint32(resultType), uint32(id), uint32(condition), uint32(op1), uint32(op2))
	return id
}

// EmitExtInst emits OpExtInst (like pack instructions) and returns the result ID.
func (b *SpvBuilder) EmitExtInst(resultType, setId SpirvId, instruction uint32, insOperands ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := append([]uint32{uint32(resultType), uint32(id), uint32(setId), instruction})
	for _, op := range insOperands {
		operands = append(operands, uint32(op))
	}
	b.instr(&b.code, spec.SpvOpExtInst, operands...)
	return id
}

// EmitExtInstImport emits OpExtInstImport and returns the result ID.
func (b *SpvBuilder) EmitExtInstImport(name string) SpirvId {
	id := b.AllocId()
	b.instr(&b.exts, spec.SpvOpExtInstImport, append([]uint32{uint32(id)}, spirvString(name)...)...)
	return id
}

// EmitUnreachable emits OpUnreachable.
func (b *SpvBuilder) EmitUnreachable() {
	b.instr(&b.code, spec.SpvOpUnreachable)
}

// Assemble combines all sections in SPIR-V specification order and returns the complete module as []uint32 ready for vkCreateShaderModule.
func (b *SpvBuilder) Assemble() []uint32 {
	var out []uint32
	out = append(out, spec.SpvMagic, spec.SpvVersion, spec.SpvGen, uint32(b.nextId), 0)
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
		if opCode == spec.SpvOpLabel {
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

// SpvBytesToWords converts a []byte SPIR-V module to []uint32 slice.
func SpvBytesToWords(bytes []byte) []uint32 {
	out := make([]uint32, len(bytes)/4)
	for i := range out {
		out[i] = uint32(bytes[i*4+0]) |
			uint32(bytes[i*4+1])<<8 |
			uint32(bytes[i*4+2])<<16 |
			uint32(bytes[i*4+3])<<24
	}

	return out
}
