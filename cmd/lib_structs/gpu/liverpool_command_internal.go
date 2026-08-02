package gpu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
)

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- liverpool_command_internal_gen.go

type LiverpoolBindPipelineInternal struct {
	// Shader addresses (for hashing).
	VertexShaderAddress   uintptr
	PixelShaderAddress    uintptr
	HullShaderAddress     uintptr
	EvalShaderAddress     uintptr
	GeometryShaderAddress uintptr

	// Draw parameters.
	PrimType uint32

	// Render target.
	RtBase         reg.GpuMemoryBase
	RtPitch        reg.CbColorPitch
	RtSlice        uint32
	RtView         reg.CbColorView
	RtAttrib       reg.CbColorAttrib
	RtTargetMask   reg.CbTargetMask
	RtColorControl reg.CbColorControl
	RtBlendControl reg.CbBlendControl
	RtClearWord0   uint32
	RtClearWord1   uint32

	// Culling and polygon mode.
	PaSuScModeCntl reg.PaSuScModeCntl

	// Render target info (decoded from CB_COLOR0_INFO).
	CbColorInfo0 reg.CbColorInfo
	CbShaderMask reg.CbShaderMask

	PaSuPolyOffsetClamp       reg.PaSuPolyOffsetClamp
	PaSuPolyOffsetFrontScale  reg.PaSuPolyOffsetFrontScale
	PaSuPolyOffsetFrontOffset reg.PaSuPolyOffsetFrontOffset
	PaSuPolyOffsetBackScale   reg.PaSuPolyOffsetBackScale
	PaSuPolyOffsetBackOffset  reg.PaSuPolyOffsetBackOffset

	// Shader control (decoded from DB_SHADER_CONTROL).
	DbShaderControl reg.DbShaderControl

	// Pixel shader input controls.
	PsInControl     reg.SpiPsInControl
	PsInputAddress  reg.SpiPsInputAddr
	PsInputControls [32]uint32

	// Shader format exports.
	SpiShaderColFormat reg.SpiShaderColFormat
	SpiShaderZFormat   reg.SpiShaderZFormat

	// Vertex shader out control.
	PaClVsOutCntl reg.PaClVsOutCntl

	// Depth buffer control.
	DbDepthControl    reg.DbDepthControl
	DbDepthSize       reg.DbDepthSize
	DbZWriteBase      reg.GpuMemoryBase
	DbZInfo           reg.DbZInfo
	DbDepthClearValue uint32

	// Stencil buffer control.
	DbStencilControl    reg.DbStencilControl
	DbStencilRefMask    reg.DbStencilrefmask
	DbStencilRefMaskBf  reg.DbStencilrefmaskBf
	DbStencilClearValue uint32

	// Render control.
	DbRenderControl reg.DbRenderControl

	// Viewport/window control.
	PaScModeCntl0         reg.PaScModeCntl0
	PaScAaConfig          reg.PaScAaConfig
	VgtMultiPrimIbResetEn reg.VgtMultiPrimIbResetEn
	PaSuLineCntl          reg.PaSuLineCntl
	PaScAaMaskX0y0X1y0    reg.PaScAaMaskX0y0X1y0
	PaScAaMaskX0y1X1y1    reg.PaScAaMaskX0y1X1y1
	MultiPrimIbResetIndex uint32

	// Hash of user data registers.
	UserDataHash uint32
}

type LiverpoolSetDynamicStateInternal struct {
	// Viewport.
	VpXScale  float32
	VpXOffset float32
	VpYScale  float32
	VpYOffset float32
	VpZScale  float32
	VpZOffset float32
	VpZMin    float32
	VpZMax    float32

	// Viewport transform engine control.
	PaClVteCntl reg.PaClVteCntl

	// Clip control.
	ClipControl reg.PaClClipCntl

	PaScModeCntl0  reg.PaScModeCntl0
	PaSuScModeCntl reg.PaSuScModeCntl

	// Guard band adjust.
	GbVertClipAdj float32
	GbVertDiscAdj float32
	GbHorzClipAdj float32
	GbHorzDiscAdj float32

	// Blend constants.
	BlendRed   uint32
	BlendGreen uint32
	BlendBlue  uint32
	BlendAlpha uint32

	// Screen scissor.
	ScissorTl reg.PaScScreenScissorTl
	ScissorBr uint32

	// Viewport scissor.
	VpScissorTl reg.PaScVportScissorTl
	VpScissorBr uint32

	// Generic scissor.
	GenericScissorTl reg.PaScGenericScissorTl
	GenericScissorBr uint32

	// Window scissor and offset.
	WindowScissorTl reg.PaScWindowScissorTl
	WindowScissorBr uint32
	WindowOffset    reg.PaScWindowOffset

	// Line stipple.
	LineStippleRepeatCount uint32
	LineStipplePattern     uint32

	// Hardware screen offset.
	HardwareScreenOffset reg.PaSuHardwareScreenOffset
}

type LiverpoolDrawInternal struct {
	// Draw parameters.
	InstanceCount uint32
	PrimType      uint32
	IsIndexed     bool

	// Indexed draw parameters.
	IndexType   uint32
	IndexBase   uintptr
	IndexCount  uint32
	IndexOffset uint32

	// Shader resources.
	VertexShRsrc1, VertexShRsrc2     uint32
	PixelShRsrc1, PixelShRsrc2       uint32
	HullShRsrc1, HullShRsrc2         uint32
	EvalShRsrc1, EvalShRsrc2         uint32
	GeometryShRsrc1, GeometryShRsrc2 uint32

	// Clear flags (from DB_RENDER_CONTROL).
	DbRenderControl     reg.DbRenderControl
	DbDepthClearValue   uint32
	DbStencilClearValue uint32

	// Hash of user data registers.
	UserDataHash uint32
}

type LiverpoolDispatchInternal struct {
	// Shader addresses (for hashing).
	ComputeShaderAddress uintptr

	// Compute parameters.
	DimX, DimY, DimZ          uint32
	ThreadX, ThreadY, ThreadZ uint32

	// Shader resources.
	ComputeShRsrc1, ComputeShRsrc2 uint32

	// Hash of user data registers.
	UserDataHash uint32
}

type LiverpoolDmaCopyInternal struct {
	SrcAddress uintptr
	DstAddress uintptr
	Count      uint32
}

type LiverpoolWaitRegMemoryInternal struct {
	Function  uint32
	Address   uintptr
	Reference uint32
	Mask      uint32
}

func (command *LiverpoolWaitRegMemoryInternal) Satisfied() bool {
	current := *(*uint32)(unsafe.Pointer(command.Address)) & command.Mask
	if ok := WaitRegMemCompare(command.Function, current, command.Reference); !ok {
		return false
	}

	return true
}

type LiverpoolWriteDataInternal struct {
	Address uintptr
	Data    []uint32
}
