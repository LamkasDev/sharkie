package spirv

import "fmt"

// EmitVariable emits a global OpVariable.
func (b *SpvBuilder) EmitVariable(ptrType, storageClass uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpVariable, ptrType, id, storageClass)
	return id
}

// EmitLocalVariable emits OpVariable with Function storage into the code section.
func (b *SpvBuilder) EmitLocalVariable(ptrType, id uint32) {
	b.instr(&b.code, SpvOpVariable, ptrType, id, SpvStorageFunction)
}

// EmitDeferredLocalVariable emits OpVariable with Function storage into the localVars section.
func (b *SpvBuilder) EmitDeferredLocalVariable(ptrType, id uint32) {
	b.instr(&b.deferredLocalVars, SpvOpVariable, ptrType, id, SpvStorageFunction)
}

// EmitPhi emits OpPhi and returns the result ID.
func (b *SpvBuilder) EmitPhi(resultType uint32, incoming ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id}, incoming...)
	b.instr(&b.code, SpvOpPhi, ops...)
	return id
}

// EmitLoadConditional emits a conditional load from a pointer.
// If cond is true, loads from ptr; otherwise returns defaultVal.
func (b *SpvBuilder) EmitLoadConditional(resultType, ptr, cond, defaultVal uint32, memoryAccess ...uint32) uint32 {
	mergeLabel := b.AllocId()
	loadLabel := b.AllocId()
	skipLabel := b.AllocId()

	// header: select branch
	b.EmitSelectionMerge(mergeLabel, SpvSelectionControlNone)
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
func (b *SpvBuilder) EmitLoad(resultType, pointer uint32, memoryAccess ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, pointer}, memoryAccess...)
	b.instr(&b.code, SpvOpLoad, ops...)
	return id
}

// EmitStore emits OpStore.
func (b *SpvBuilder) EmitStore(pointer, object uint32, memoryAccess ...uint32) {
	if pointer == 0 {
		panic(fmt.Sprintf("id is zero"))
	}
	ops := append([]uint32{pointer, object}, memoryAccess...)
	b.instr(&b.code, SpvOpStore, ops...)
}

// EmitMemoryBarrier emits OpMemoryBarrier with the specified scope and semantics.
func (b *SpvBuilder) EmitMemoryBarrier(scope, semantics uint32) {
	b.instr(&b.code, SpvOpMemoryBarrier, scope, semantics)
}
