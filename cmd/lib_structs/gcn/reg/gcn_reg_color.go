package reg

type CbColorInfo Reg

func (r CbColorInfo) Endian() uint32               { return Reg(r).Extract(0, 0x3) } // TODO: Unused in pipeline
func (r CbColorInfo) Format() uint32               { return Reg(r).Extract(2, 0x1F) }
func (r CbColorInfo) LinearGeneral() bool          { return Reg(r).ExtractBool(7) } // TODO: Unused in pipeline
func (r CbColorInfo) NumberType() uint32           { return Reg(r).Extract(8, 0x7) }
func (r CbColorInfo) CompSwap() uint32             { return Reg(r).Extract(11, 0x3) }
func (r CbColorInfo) FastClear() bool              { return Reg(r).ExtractBool(13) }
func (r CbColorInfo) Compression() bool            { return Reg(r).ExtractBool(14) }
func (r CbColorInfo) BlendClamp() bool             { return Reg(r).ExtractBool(15) } // TODO: Unused in pipeline
func (r CbColorInfo) BlendBypass() bool            { return Reg(r).ExtractBool(16) }
func (r CbColorInfo) SimpleFloat() bool            { return Reg(r).ExtractBool(17) }  // TODO: Unused in pipeline
func (r CbColorInfo) RoundMode() uint32            { return Reg(r).Extract(18, 1) }   // TODO: Unused in pipeline
func (r CbColorInfo) CmaskIsLinear() bool          { return Reg(r).ExtractBool(19) }  // TODO: Unused in pipeline
func (r CbColorInfo) BlendOptDontRdDst() uint32    { return Reg(r).Extract(20, 0x7) } // TODO: Unused in pipeline
func (r CbColorInfo) BlendOptDiscardPixel() uint32 { return Reg(r).Extract(23, 0x7) } // TODO: Unused in pipeline
func (r CbColorInfo) FmaskCompressionDis() bool    { return Reg(r).ExtractBool(26) }  // TODO: Unused in pipeline

type CbColorControl Reg

func (r CbColorControl) DegammaEnable() bool { return Reg(r).ExtractBool(3) } // TODO: Unused in pipeline
func (r CbColorControl) Mode() uint32        { return Reg(r).Extract(4, 0x7) }
func (r CbColorControl) Rop3() uint32        { return Reg(r).Extract(16, 0xFF) }

type CbBlendControl Reg

func (r CbBlendControl) ColorSrcblend() uint32    { return Reg(r).Extract(0, 0x1F) }
func (r CbBlendControl) ColorCombFcn() uint32     { return Reg(r).Extract(5, 0x7) }
func (r CbBlendControl) ColorDestblend() uint32   { return Reg(r).Extract(8, 0x1F) }
func (r CbBlendControl) AlphaSrcblend() uint32    { return Reg(r).Extract(16, 0x1F) }
func (r CbBlendControl) AlphaCombFcn() uint32     { return Reg(r).Extract(21, 0x7) }
func (r CbBlendControl) AlphaDestblend() uint32   { return Reg(r).Extract(24, 0x1F) }
func (r CbBlendControl) SeparateAlphaBlend() bool { return Reg(r).ExtractBool(29) }
func (r CbBlendControl) Enable() bool             { return Reg(r).ExtractBool(30) }
func (r CbBlendControl) DisableRop3() bool        { return Reg(r).ExtractBool(31) }

type CbTargetMask Reg

func (r CbTargetMask) Target0Enable() uint32 { return Reg(r).Extract(0, 0xF) }
func (r CbTargetMask) Target1Enable() uint32 { return Reg(r).Extract(4, 0xF) }
func (r CbTargetMask) Target2Enable() uint32 { return Reg(r).Extract(8, 0xF) }
func (r CbTargetMask) Target3Enable() uint32 { return Reg(r).Extract(12, 0xF) }
func (r CbTargetMask) Target4Enable() uint32 { return Reg(r).Extract(16, 0xF) }
func (r CbTargetMask) Target5Enable() uint32 { return Reg(r).Extract(20, 0xF) }
func (r CbTargetMask) Target6Enable() uint32 { return Reg(r).Extract(24, 0xF) }
func (r CbTargetMask) Target7Enable() uint32 { return Reg(r).Extract(28, 0xF) }

type CbColorPitch Reg

func (r CbColorPitch) TileMax() uint32      { return Reg(r).Extract(0, 0x7FF) }  // TODO: Unused in pipeline
func (r CbColorPitch) FmaskTileMax() uint32 { return Reg(r).Extract(20, 0x7FF) } // TODO: Unused in pipeline

type CbColorView Reg

func (r CbColorView) SliceStart() uint32 { return Reg(r).Extract(0, 0x7FF) }  // TODO: Unused in pipeline
func (r CbColorView) SliceMax() uint32   { return Reg(r).Extract(13, 0x7FF) } // TODO: Unused in pipeline

type CbColorAttrib Reg

func (r CbColorAttrib) TileModeIndex() uint32      { return Reg(r).Extract(0, 0x1F) } // TODO: Unused in pipeline
func (r CbColorAttrib) FmaskTileModeIndex() uint32 { return Reg(r).Extract(5, 0x1F) } // TODO: Unused in pipeline
func (r CbColorAttrib) NumSamples() uint32         { return Reg(r).Extract(12, 0x7) } // TODO: Unused in pipeline
func (r CbColorAttrib) NumFragments() uint32       { return Reg(r).Extract(15, 0x3) } // TODO: Unused in pipeline
func (r CbColorAttrib) ForceDstAlpha1() bool       { return Reg(r).ExtractBool(17) }  // TODO: Unused in pipeline

type CbShaderMask Reg

func (r CbShaderMask) Output0Enable() uint32 { return Reg(r).Extract(0, 0xF) }  // TODO: Unused in pipeline
func (r CbShaderMask) Output1Enable() uint32 { return Reg(r).Extract(4, 0xF) }  // TODO: Unused in pipeline
func (r CbShaderMask) Output2Enable() uint32 { return Reg(r).Extract(8, 0xF) }  // TODO: Unused in pipeline
func (r CbShaderMask) Output3Enable() uint32 { return Reg(r).Extract(12, 0xF) } // TODO: Unused in pipeline
func (r CbShaderMask) Output4Enable() uint32 { return Reg(r).Extract(16, 0xF) } // TODO: Unused in pipeline
func (r CbShaderMask) Output5Enable() uint32 { return Reg(r).Extract(20, 0xF) } // TODO: Unused in pipeline
func (r CbShaderMask) Output6Enable() uint32 { return Reg(r).Extract(24, 0xF) } // TODO: Unused in pipeline
func (r CbShaderMask) Output7Enable() uint32 { return Reg(r).Extract(28, 0xF) } // TODO: Unused in pipeline
