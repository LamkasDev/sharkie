package common

import (
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

// EmitControlBarrier emits OpControlBarrier.
func (b *SpvBuilder) EmitControlBarrier(executionScope, memoryScope, semantics SpirvId) {
	b.instr(&b.code, spec.SpvOpControlBarrier, uint32(executionScope), uint32(memoryScope), uint32(semantics))
}

// EmitMemoryBarrier emits OpMemoryBarrier with the specified scope and semantics.
func (b *SpvBuilder) EmitMemoryBarrier(scope, semantics SpirvId) {
	b.instr(&b.code, spec.SpvOpMemoryBarrier, uint32(scope), uint32(semantics))
}
