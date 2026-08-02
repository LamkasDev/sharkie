package reg

type DbShaderControl Reg

func (r DbShaderControl) ZExportEnable() bool              { return Reg(r).ExtractBool(0) } // TODO: Unused in pipeline
func (r DbShaderControl) StencilTestValExportEnable() bool { return Reg(r).ExtractBool(1) } // TODO: Unused in pipeline
func (r DbShaderControl) StencilOpValExportEnable() bool   { return Reg(r).ExtractBool(2) } // TODO: Unused in pipeline
func (r DbShaderControl) ZOrder() uint32                   { return Reg(r).Extract(4, 0x3) }
func (r DbShaderControl) KillEnable() bool                 { return Reg(r).ExtractBool(6) }
func (r DbShaderControl) CoverageToMaskEnable() bool       { return Reg(r).ExtractBool(7) }
func (r DbShaderControl) MaskExportEnable() bool           { return Reg(r).ExtractBool(8) }  // TODO: Unused in pipeline
func (r DbShaderControl) ExecOnHierFail() bool             { return Reg(r).ExtractBool(9) }  // TODO: Unused in pipeline
func (r DbShaderControl) ExecOnNoop() bool                 { return Reg(r).ExtractBool(10) } // TODO: Unused in pipeline
func (r DbShaderControl) AlphaToMaskDisable() bool         { return Reg(r).ExtractBool(11) }
func (r DbShaderControl) DepthBeforeShader() bool          { return Reg(r).ExtractBool(12) }
func (r DbShaderControl) ConservativeZExport() uint32      { return Reg(r).Extract(13, 0x3) } // TODO: Unused in pipeline

type DbRenderControl Reg

func (r DbRenderControl) DepthClearEnable() bool       { return Reg(r).ExtractBool(0) }
func (r DbRenderControl) StencilClearEnable() bool     { return Reg(r).ExtractBool(1) }
func (r DbRenderControl) DepthCopy() bool              { return Reg(r).ExtractBool(2) } // TODO: Unused in pipeline
func (r DbRenderControl) StencilCopy() bool            { return Reg(r).ExtractBool(3) } // TODO: Unused in pipeline
func (r DbRenderControl) ResummarizeEnable() bool      { return Reg(r).ExtractBool(4) } // TODO: Unused in pipeline
func (r DbRenderControl) StencilCompressDisable() bool { return Reg(r).ExtractBool(5) } // TODO: Unused in pipeline
func (r DbRenderControl) DepthCompressDisable() bool   { return Reg(r).ExtractBool(6) }
func (r DbRenderControl) CopyCentroid() bool           { return Reg(r).ExtractBool(7) }  // TODO: Unused in pipeline
func (r DbRenderControl) CopySample() uint32           { return Reg(r).Extract(8, 0xF) } // TODO: Unused in pipeline

type DbZInfo Reg

func (r DbZInfo) Format() uint32          { return Reg(r).Extract(0, 0x3) }
func (r DbZInfo) NumSamples() uint32      { return Reg(r).Extract(2, 0x3) }  // TODO: Unused in pipeline
func (r DbZInfo) TileSplit() uint32       { return Reg(r).Extract(13, 0x7) } // TODO: Unused in pipeline
func (r DbZInfo) AllowExpclear() bool     { return Reg(r).ExtractBool(27) }  // TODO: Unused in pipeline
func (r DbZInfo) ReadSize() bool          { return Reg(r).ExtractBool(28) }  // TODO: Unused in pipeline
func (r DbZInfo) TileSurfaceEnable() bool { return Reg(r).ExtractBool(29) }  // TODO: Unused in pipeline
func (r DbZInfo) ZrangePrecision() bool   { return Reg(r).ExtractBool(31) }  // TODO: Unused in pipeline

type DbDepthControl Reg

func (r DbDepthControl) StencilEnable() bool                 { return Reg(r).ExtractBool(0) }
func (r DbDepthControl) ZEnable() bool                       { return Reg(r).ExtractBool(1) }
func (r DbDepthControl) ZWriteEnable() bool                  { return Reg(r).ExtractBool(2) }
func (r DbDepthControl) DepthBoundsEnable() bool             { return Reg(r).ExtractBool(3) }
func (r DbDepthControl) Zfunc() uint32                       { return Reg(r).Extract(4, 0x7) }
func (r DbDepthControl) BackfaceEnable() bool                { return Reg(r).ExtractBool(7) }
func (r DbDepthControl) Stencilfunc() uint32                 { return Reg(r).Extract(8, 0x7) }
func (r DbDepthControl) StencilfuncBf() uint32               { return Reg(r).Extract(20, 0x7) }
func (r DbDepthControl) EnableColorWritesOnDepthFail() bool  { return Reg(r).ExtractBool(30) } // TODO: Unused in pipeline
func (r DbDepthControl) DisableColorWritesOnDepthPass() bool { return Reg(r).ExtractBool(31) } // TODO: Unused in pipeline

type DbStencilControl Reg

func (r DbStencilControl) Stencilfail() uint32    { return Reg(r).Extract(0, 0xF) }
func (r DbStencilControl) Stencilzpass() uint32   { return Reg(r).Extract(4, 0xF) }
func (r DbStencilControl) Stencilzfail() uint32   { return Reg(r).Extract(8, 0xF) }
func (r DbStencilControl) StencilfailBf() uint32  { return Reg(r).Extract(12, 0xF) }
func (r DbStencilControl) StencilzpassBf() uint32 { return Reg(r).Extract(16, 0xF) }
func (r DbStencilControl) StencilzfailBf() uint32 { return Reg(r).Extract(20, 0xF) }

type DbDepthSize Reg

func (r DbDepthSize) PitchTileMax() uint32  { return Reg(r).Extract(0, 0x7FF) }  // TODO: Unused in pipeline
func (r DbDepthSize) HeightTileMax() uint32 { return Reg(r).Extract(11, 0x7FF) } // TODO: Unused in pipeline

type DbStencilrefmask Reg

func (r DbStencilrefmask) Stenciltestval() uint32   { return Reg(r).Extract(0, 0xFF) }
func (r DbStencilrefmask) Stencilmask() uint32      { return Reg(r).Extract(8, 0xFF) }
func (r DbStencilrefmask) Stencilwritemask() uint32 { return Reg(r).Extract(16, 0xFF) }
func (r DbStencilrefmask) Stencilopval() uint32     { return Reg(r).Extract(24, 0xFF) } // TODO: Unused in pipeline

type DbStencilrefmaskBf Reg

func (r DbStencilrefmaskBf) StenciltestvalBf() uint32   { return Reg(r).Extract(0, 0xFF) }
func (r DbStencilrefmaskBf) StencilmaskBf() uint32      { return Reg(r).Extract(8, 0xFF) }
func (r DbStencilrefmaskBf) StencilwritemaskBf() uint32 { return Reg(r).Extract(16, 0xFF) }
func (r DbStencilrefmaskBf) StencilopvalBf() uint32     { return Reg(r).Extract(24, 0xFF) } // TODO: Unused in pipeline
