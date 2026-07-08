package gpu

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- liverpool_command_internal_gen.go

type LiverpoolBindPipelineInternal struct {
	// Shader addresses (for hashing).
	VertexShaderAddress   uintptr
	HullShaderAddress     uintptr
	EvalShaderAddress     uintptr
	GeometryShaderAddress uintptr
	PixelShaderAddress    uintptr

	// Draw parameters.
	PrimType uint32

	// Render target.
	RtBase         uint32
	RtPitch        uint32
	RtSlice        uint32
	RtView         uint32
	RtAttrib       uint32
	RtTargetMask   uint32
	RtColorControl uint32
	RtBlendControl uint32
	RtClearWord0   uint32
	RtClearWord1   uint32

	// Culling and polygon mode.
	CullFront             bool
	CullBack              bool
	Face                  bool
	PolyMode              uint32
	PolyModeFrontPtype    uint32
	PolyModeBackPtype     uint32
	PolyOffsetFrontEnable bool
	PolyOffsetBackEnable  bool
	PolyOffsetParaEnable  bool
	ProvokingVertexLast   bool

	// Render target info (decoded from CB_COLOR0_INFO).
	RtFormat               uint32
	RtNumberType           uint32
	RtCompSwap             uint32
	RtLinearGeneral        bool
	RtFastClear            bool
	RtCompression          bool
	RtBlendClamp           bool
	RtBlendBypass          bool
	RtSimpleFloat          bool
	RtRoundMode            uint32
	RtCmaskIsLinear        bool
	RtBlendOptDontRdDst    uint32
	RtBlendOptDiscardPixel uint32
	RtFmaskCompressionDis  bool

	// Shader control (decoded from DB_SHADER_CONTROL).
	DbZExportEnable              bool
	DbStencilTestValExportEnable bool
	DbStencilOpValExportEnable   bool
	DbZOrder                     uint32
	DbKillEnable                 bool
	DbCoverageToMaskEnable       bool
	DbMaskExportEnable           bool
	DbExecOnHierFail             bool
	DbExecOnNoop                 bool
	DbAlphaToMaskDisable         bool
	DbDepthBeforeShader          bool
	DbConservativeZExport        uint32

	// Pixel shader input controls.
	PsInControl     uint32
	PsInputAddress  uint32
	PsInputControls [32]uint32

	// Depth buffer control.
	DbDepthControl    uint32
	DbDepthSize       uint32
	DbZWriteBase      uint32
	DbZFormat         uint32
	DbDepthClearValue uint32

	// Stencil buffer control.
	DbStencilControl    uint32
	DbStencilRefMask    uint32
	DbStencilRefMaskBf  uint32
	DbStencilClearValue uint32

	// Render control.
	DbDepthClearEnable   bool
	DbStencilClearEnable bool
	DbDepthCopy          bool
	DbStencilCopy        bool

	// Viewport/window control.
	VpScissorEnable    bool
	WindowOffsetEnable bool

	// Line stipple.
	LineStippleEnable      bool
	LineStippleRepeatCount uint32
	LineStipplePattern     uint32

	// Anti-aliasing control.
	MsaaEnable          bool
	MsaaSampleLocations uint32

	// Multi primitive index buffer reset.
	MultiPrimIbResetEnable bool
	MultiPrimIbResetIndex  uint32
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
	VteControl      uint32
	VpXScaleEnable  bool
	VpXOffsetEnable bool
	VpYScaleEnable  bool
	VpYOffsetEnable bool
	VpZScaleEnable  bool
	VpZOffsetEnable bool
	VtxXyFmt        bool
	VtxZFmt         bool
	VtxW0Fmt        bool

	// Clip control.
	ClipControl uint32

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
	ScissorTl uint32
	ScissorBr uint32

	// Viewport scissor.
	VpScissorEnable bool
	VpScissorTl     uint32
	VpScissorBr     uint32

	// Generic scissor.
	GenericScissorTl uint32
	GenericScissorBr uint32

	// Window scissor and offset.
	WindowOffsetEnable bool
	WindowScissorTl    uint32
	WindowScissorBr    uint32
	WindowOffset       uint32

	// Line stipple.
	LineStippleRepeatCount uint32
	LineStipplePattern     uint32

	// Hardware screen offset.
	HardwareScreenOffset uint32
}

type LiverpoolDrawInternal struct {
	// Draw parameters.
	VertexCount   uint32
	InstanceCount uint32
	PrimType      uint32
	IsIndexed     bool

	// Indexed draw parameters.
	IndexCount       uint32
	IndexType        uint32
	IndexBaseAddress uintptr
	IndexOffset      uint32

	// Shader resources.
	VertexShRsrc1, VertexShRsrc2     uint32
	PixelShRsrc1, PixelShRsrc2       uint32
	HullShRsrc1, HullShRsrc2         uint32
	EvalShRsrc1, EvalShRsrc2         uint32
	GeometryShRsrc1, GeometryShRsrc2 uint32

	// Clear flags (from DB_RENDER_CONTROL).
	DbDepthClearEnable   bool
	DbStencilClearEnable bool
	DbDepthClearValue    uint32
	DbStencilClearValue  uint32

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
