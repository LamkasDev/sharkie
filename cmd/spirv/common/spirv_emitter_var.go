package common

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

// EmitVariable emits a global OpVariable.
func (b *SpvBuilder) EmitVariable(ptrType SpirvId, storageClass uint32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpVariable, uint32(ptrType), uint32(id), storageClass)
	return id
}

// EmitLocalVariable emits OpVariable with Function storage into the code section.
func (b *SpvBuilder) EmitLocalVariable(ptrType, id SpirvId) {
	b.instr(&b.code, spec.SpvOpVariable, uint32(ptrType), uint32(id), spec.SpvStorageFunction)
}

// EmitDeferredLocalVariable emits OpVariable with Function storage into the localVars section.
func (b *SpvBuilder) EmitDeferredLocalVariable(ptrType, id SpirvId) {
	b.instr(&b.deferredLocalVars, spec.SpvOpVariable, uint32(ptrType), uint32(id), spec.SpvStorageFunction)
}

// EmitPhi emits OpPhi and returns the result ID.
func (b *SpvBuilder) EmitPhi(resultType SpirvId, incoming ...SpirvId) SpirvId {
	id := b.AllocId()
	ops := []uint32{uint32(resultType), uint32(id)}
	for _, in := range incoming {
		ops = append(ops, uint32(in))
	}
	b.instr(&b.code, spec.SpvOpPhi, ops...)
	return id
}

// EmitLoadConditional emits a conditional load from a pointer.
// If cond is true, loads from ptr; otherwise returns defaultVal.
func (b *SpvBuilder) EmitLoadConditional(resultType, ptr, cond, defaultVal SpirvId, memoryAccess ...uint32) SpirvId {
	mergeLabel := b.AllocId()
	loadLabel := b.AllocId()
	skipLabel := b.AllocId()

	// header: select branch
	b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(cond, loadLabel, skipLabel)

	// load block
	b.EmitLabel(loadLabel)
	loaded := b.EmitLoad(resultType, ptr, memoryAccess...)
	b.EmitBranch(mergeLabel)

	// skip block
	b.EmitLabel(skipLabel)
	b.EmitBranch(mergeLabel)

	// merge + phi
	b.EmitLabel(mergeLabel)
	return b.EmitPhi(resultType, loaded, loadLabel, defaultVal, skipLabel)
}

// EmitLoad emits OpLoad and returns the result ID.
func (b *SpvBuilder) EmitLoad(resultType, pointer SpirvId, memoryAccess ...uint32) SpirvId {
	id := b.AllocId()
	ops := append([]uint32{uint32(resultType), uint32(id), uint32(pointer)}, memoryAccess...)
	b.instr(&b.code, spec.SpvOpLoad, ops...)
	return id
}

// EmitStore emits OpStore.
func (b *SpvBuilder) EmitStore(pointer, object SpirvId, memoryAccess ...uint32) {
	if pointer == 0 {
		panic(fmt.Sprintf("id is zero"))
	}
	ops := append([]uint32{uint32(pointer), uint32(object)}, memoryAccess...)
	b.instr(&b.code, spec.SpvOpStore, ops...)
}
