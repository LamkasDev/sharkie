package common

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
)

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- spirv_emitter_shader_context_gen.go

type SpirvShaderContext interface{}

type SpirvVertexShaderContext struct {
	ClipDistEnable          uint8
	CullDistEnable          uint8
	FetchShaderAddress      uintptr
	FetchShaderInstructions []*gcnSpec.Instruction `hsp:"-"`
}

type SpirvFragmentShaderContext struct {
	PsInControl       uint32
	PsInputAddress    uint32
	PsInputControls   [32]uint32
	DepthBeforeShader bool
	ZOrder            uint32
	FrontFaceEnable   bool
}

type SpirvComputeShaderContext struct {
	ThreadX uint32
	ThreadY uint32
	ThreadZ uint32
}
