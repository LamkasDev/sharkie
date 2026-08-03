package reg

type SpiPsInControl Reg

func (r SpiPsInControl) NumInterp() uint32         { return Reg(r).Extract(0, 0x3F) } // TODO: Unused in pipeline
func (r SpiPsInControl) ParamGen() bool            { return Reg(r).ExtractBool(6) }   // TODO: Unused in pipeline
func (r SpiPsInControl) BcOptimizeDisable() uint32 { return Reg(r).Extract(14, 0x3) } // TODO: Unused in pipeline

type SpiPsInputAddr Reg

func (r SpiPsInputAddr) PerspSampleEna() bool    { return Reg(r).ExtractBool(0) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PerspCenterEna() bool    { return Reg(r).ExtractBool(1) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PerspCentroidEna() bool  { return Reg(r).ExtractBool(2) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PerspPullModelEna() bool { return Reg(r).ExtractBool(3) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) LinearSampleEna() bool   { return Reg(r).ExtractBool(4) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) LinearCenterEna() bool   { return Reg(r).ExtractBool(5) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) LinearCentroidEna() bool { return Reg(r).ExtractBool(6) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) LineStippleTexEna() bool { return Reg(r).ExtractBool(7) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PosXFloatEna() bool      { return Reg(r).ExtractBool(8) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PosYFloatEna() bool      { return Reg(r).ExtractBool(9) }  // TODO: Unused in pipeline
func (r SpiPsInputAddr) PosZFloatEna() bool      { return Reg(r).ExtractBool(10) } // TODO: Unused in pipeline
func (r SpiPsInputAddr) PosWFloatEna() bool      { return Reg(r).ExtractBool(11) } // TODO: Unused in pipeline
func (r SpiPsInputAddr) FrontFaceEna() bool      { return Reg(r).ExtractBool(12) }
func (r SpiPsInputAddr) AncillaryEna() bool      { return Reg(r).ExtractBool(13) } // TODO: Unused in pipeline
func (r SpiPsInputAddr) SampleCoverageEna() bool { return Reg(r).ExtractBool(14) } // TODO: Unused in pipeline
func (r SpiPsInputAddr) PosFixedPtEna() bool     { return Reg(r).ExtractBool(15) } // TODO: Unused in pipeline

type SpiShaderColFormat uint32

func (r SpiShaderColFormat) Col0ExportFormat() uint32 { return Reg(r).Extract(0, 0xF) }
func (r SpiShaderColFormat) Col1ExportFormat() uint32 { return Reg(r).Extract(4, 0xF) }
func (r SpiShaderColFormat) Col2ExportFormat() uint32 { return Reg(r).Extract(8, 0xF) }
func (r SpiShaderColFormat) Col3ExportFormat() uint32 { return Reg(r).Extract(12, 0xF) }
func (r SpiShaderColFormat) Col4ExportFormat() uint32 { return Reg(r).Extract(16, 0xF) }
func (r SpiShaderColFormat) Col5ExportFormat() uint32 { return Reg(r).Extract(20, 0xF) }
func (r SpiShaderColFormat) Col6ExportFormat() uint32 { return Reg(r).Extract(24, 0xF) }
func (r SpiShaderColFormat) Col7ExportFormat() uint32 { return Reg(r).Extract(28, 0xF) }

type SpiShaderZFormat uint32

func (r SpiShaderZFormat) ZExportFormat() uint32 { return Reg(r).Extract(0, 0xF) } // TODO: Unused in pipeline

type SpiVsOutConfig Reg

func (r SpiVsOutConfig) VsExportCount() uint32 { return Reg(r).Extract(1, 0x1F) }
func (r SpiVsOutConfig) VsHalfPack() bool      { return Reg(r).ExtractBool(6) } // TODO: Unused in pipeline
