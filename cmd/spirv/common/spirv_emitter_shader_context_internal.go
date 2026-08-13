package common

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
)

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- spirv_emitter_shader_context_internal_gen.go

type SpirvShaderKey struct {
	Address     uintptr
	ContextHash uint64
}

type SpirvShaderContext interface{}

type SpirvVertexShaderContext struct {
	ClipDistEnable uint8
	CullDistEnable uint8
	VsExportCount  uint32
	MubufFormats   []SpirvMubufFormat

	FetchShaderAddress      uintptr
	FetchShaderInstructions []*gcnSpec.Instruction `hsp:"-"`
}

type SpirvFragmentShaderContext struct {
	PsInControl     uint32
	PsInputAddress  uint32
	PsInputControls [32]uint32
	MubufFormats    []SpirvMubufFormat

	DepthBeforeShader bool
	ZOrder            uint32
	FrontFaceEnable   bool
}

type SpirvComputeShaderContext struct {
	ThreadX      uint32
	ThreadY      uint32
	ThreadZ      uint32
	MubufFormats []SpirvMubufFormat
}

type SpirvMubufFormat struct {
	PC         uint32
	DataFormat uint8
	NumFormat  uint8
}
