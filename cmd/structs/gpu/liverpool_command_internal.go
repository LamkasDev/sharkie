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

	// Render target info fields (decoded from CB_COLOR0_INFO).
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

	// Shader control fields (decoded from DB_SHADER_CONTROL).
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

	// Depth buffer control.
	DbDepthControl    uint32
	DbDepthSize       uint32
	DbZWriteBase      uint32
	DbZFormat         uint32
	DbDepthClearValue uint32

	// Stencil buffer control.
	DbStencilControl    uint32
	DbStencilClearValue uint32

	// Render control.
	DbDepthClearEnable   bool
	DbStencilClearEnable bool
	DbDepthCopy          bool
	DbStencilCopy        bool
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

	// Blend constants.
	BlendRed   uint32
	BlendGreen uint32
	BlendBlue  uint32
	BlendAlpha uint32

	// Screen scissor.
	ScissorTl uint32
	ScissorBr uint32
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
	BaseVertexOffset uint32

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
