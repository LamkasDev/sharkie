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

// EmitLoad emits OpLoad and returns the result ID.
func (b *SpvBuilder) EmitLoad(resultType, pointer uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpLoad, resultType, id, pointer)
	return id
}

// EmitStore emits OpStore.
func (b *SpvBuilder) EmitStore(pointer, object uint32) {
	if pointer == 0 {
		panic(fmt.Sprintf("id is zero"))
	}
	b.instr(&b.code, SpvOpStore, pointer, object)
}
