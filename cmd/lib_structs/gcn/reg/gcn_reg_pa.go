package reg

import "math"

type PaSuScModeCntl Reg

func (r PaSuScModeCntl) CullFront() bool             { return Reg(r).ExtractBool(0) }
func (r PaSuScModeCntl) CullBack() bool              { return Reg(r).ExtractBool(1) }
func (r PaSuScModeCntl) Face() bool                  { return Reg(r).ExtractBool(2) }
func (r PaSuScModeCntl) PolyMode() uint32            { return Reg(r).Extract(3, 0x3) }
func (r PaSuScModeCntl) PolyModeFrontPtype() uint32  { return Reg(r).Extract(5, 0x7) }
func (r PaSuScModeCntl) PolyModeBackPtype() uint32   { return Reg(r).Extract(8, 0x7) } // TODO: Unused in pipeline
func (r PaSuScModeCntl) PolyOffsetFrontEnable() bool { return Reg(r).ExtractBool(11) }
func (r PaSuScModeCntl) PolyOffsetBackEnable() bool  { return Reg(r).ExtractBool(12) }
func (r PaSuScModeCntl) PolyOffsetParaEnable() bool  { return Reg(r).ExtractBool(13) } // TODO: Unused in pipeline
func (r PaSuScModeCntl) WindowOffsetEnable() bool    { return Reg(r).ExtractBool(16) } // TODO: Unused in pipeline
func (r PaSuScModeCntl) ProvokingVertexLast() bool   { return Reg(r).ExtractBool(19) }
func (r PaSuScModeCntl) PerspCorrDis() bool          { return Reg(r).ExtractBool(20) } // TODO: Unused in pipeline
func (r PaSuScModeCntl) MultiPrimIbEna() bool        { return Reg(r).ExtractBool(21) } // TODO: Unused in pipeline

type PaScModeCntl0 Reg

func (r PaScModeCntl0) MsaaEnable() bool           { return Reg(r).ExtractBool(0) } // TODO: Unused in pipeline
func (r PaScModeCntl0) VpScissorEnable() bool      { return Reg(r).ExtractBool(1) }
func (r PaScModeCntl0) LineStippleEnable() bool    { return Reg(r).ExtractBool(2) }
func (r PaScModeCntl0) SendUnlitStilesToPkr() bool { return Reg(r).ExtractBool(3) } // TODO: Unused in pipeline

type PaClVteCntl Reg

func (r PaClVteCntl) VpXScaleEnable() bool  { return Reg(r).ExtractBool(0) }
func (r PaClVteCntl) VpXOffsetEnable() bool { return Reg(r).ExtractBool(1) }
func (r PaClVteCntl) VpYScaleEnable() bool  { return Reg(r).ExtractBool(2) }
func (r PaClVteCntl) VpYOffsetEnable() bool { return Reg(r).ExtractBool(3) }
func (r PaClVteCntl) VpZScaleEnable() bool  { return Reg(r).ExtractBool(4) }
func (r PaClVteCntl) VpZOffsetEnable() bool { return Reg(r).ExtractBool(5) }
func (r PaClVteCntl) VtxXyFmt() bool        { return Reg(r).ExtractBool(8) }  // TODO: Unused in pipeline
func (r PaClVteCntl) VtxZFmt() bool         { return Reg(r).ExtractBool(9) }  // TODO: Unused in pipeline
func (r PaClVteCntl) VtxW0Fmt() bool        { return Reg(r).ExtractBool(10) } // TODO: Unused in pipeline

type PaSuLineStippleCntl Reg

func (r PaSuLineStippleCntl) LineStippleReset() uint32       { return Reg(r).Extract(0, 0x3) }    // TODO: Unused in pipeline
func (r PaSuLineStippleCntl) ExpandFullLength() bool         { return Reg(r).ExtractBool(2) }     // TODO: Unused in pipeline
func (r PaSuLineStippleCntl) FractionalAccum() bool          { return Reg(r).ExtractBool(3) }     // TODO: Unused in pipeline
func (r PaSuLineStippleCntl) DiamondAdjust() bool            { return Reg(r).ExtractBool(4) }     // TODO: Unused in pipeline
func (r PaSuLineStippleCntl) LineStippleRepeatCount() uint32 { return Reg(r).Extract(16, 0xFF) }  // TODO: Unused in pipeline
func (r PaSuLineStippleCntl) LineStipplePattern() uint32     { return Reg(r).Extract(0, 0xFFFF) } // TODO: Unused in pipeline

type PaClVportFloat Reg

func (r PaClVportFloat) Float32() float32 { return math.Float32frombits(uint32(r)) }

type PaScAaConfig Reg

func (r PaScAaConfig) MsaaNumSamples() uint32      { return Reg(r).Extract(0, 0x7) }  // TODO: Unused in pipeline
func (r PaScAaConfig) AaMaskCentroidDtmn() bool    { return Reg(r).ExtractBool(4) }   // TODO: Unused in pipeline
func (r PaScAaConfig) MaxSampleDist() uint32       { return Reg(r).Extract(13, 0xF) } // TODO: Unused in pipeline
func (r PaScAaConfig) MsaaExposedSamples() uint32  { return Reg(r).Extract(20, 0x7) } // TODO: Unused in pipeline
func (r PaScAaConfig) DetailToExposedMode() uint32 { return Reg(r).Extract(24, 0x3) } // TODO: Unused in pipeline

type PaClClipCntl Reg

func (r PaClClipCntl) UcpEna0() bool                { return Reg(r).ExtractBool(0) }   // TODO: Unused in pipeline
func (r PaClClipCntl) UcpEna1() bool                { return Reg(r).ExtractBool(1) }   // TODO: Unused in pipeline
func (r PaClClipCntl) UcpEna2() bool                { return Reg(r).ExtractBool(2) }   // TODO: Unused in pipeline
func (r PaClClipCntl) UcpEna3() bool                { return Reg(r).ExtractBool(3) }   // TODO: Unused in pipeline
func (r PaClClipCntl) UcpEna4() bool                { return Reg(r).ExtractBool(4) }   // TODO: Unused in pipeline
func (r PaClClipCntl) UcpEna5() bool                { return Reg(r).ExtractBool(5) }   // TODO: Unused in pipeline
func (r PaClClipCntl) PsUcpYScaleNeg() bool         { return Reg(r).ExtractBool(13) }  // TODO: Unused in pipeline
func (r PaClClipCntl) PsUcpMode() uint32            { return Reg(r).Extract(14, 0x3) } // TODO: Unused in pipeline
func (r PaClClipCntl) ClipDisable() bool            { return Reg(r).ExtractBool(16) }  // TODO: Unused in pipeline
func (r PaClClipCntl) UcpCullOnlyEna() bool         { return Reg(r).ExtractBool(17) }  // TODO: Unused in pipeline
func (r PaClClipCntl) BoundaryEdgeFlagEna() bool    { return Reg(r).ExtractBool(18) }  // TODO: Unused in pipeline
func (r PaClClipCntl) DxClipSpaceDef() bool         { return Reg(r).ExtractBool(19) }  // TODO: Unused in pipeline
func (r PaClClipCntl) DisClipErrDetect() bool       { return Reg(r).ExtractBool(20) }  // TODO: Unused in pipeline
func (r PaClClipCntl) VtxKillOr() bool              { return Reg(r).ExtractBool(21) }  // TODO: Unused in pipeline
func (r PaClClipCntl) DxRasterizationKill() bool    { return Reg(r).ExtractBool(22) }  // TODO: Unused in pipeline
func (r PaClClipCntl) DxLinearAttrClipEna() bool    { return Reg(r).ExtractBool(24) }  // TODO: Unused in pipeline
func (r PaClClipCntl) VteVportProvokeDisable() bool { return Reg(r).ExtractBool(25) }  // TODO: Unused in pipeline
func (r PaClClipCntl) ZclipNearDisable() bool       { return Reg(r).ExtractBool(26) }  // TODO: Unused in pipeline
func (r PaClClipCntl) ZclipFarDisable() bool        { return Reg(r).ExtractBool(27) }  // TODO: Unused in pipeline

type PaScScreenScissorTl Reg

func (r PaScScreenScissorTl) TlX() uint32 { return Reg(r).Extract(0, 0xFFFF) }
func (r PaScScreenScissorTl) TlY() uint32 { return Reg(r).Extract(16, 0xFFFF) }

type PaScVportScissorTl Reg

func (r PaScVportScissorTl) TlX() uint32               { return Reg(r).Extract(0, 0x7FFF) }  // TODO: Unused in pipeline
func (r PaScVportScissorTl) TlY() uint32               { return Reg(r).Extract(16, 0x7FFF) } // TODO: Unused in pipeline
func (r PaScVportScissorTl) WindowOffsetDisable() bool { return Reg(r).ExtractBool(31) }     // TODO: Unused in pipeline

type PaScGenericScissorTl Reg

func (r PaScGenericScissorTl) TlX() uint32               { return Reg(r).Extract(0, 0x7FFF) }  // TODO: Unused in pipeline
func (r PaScGenericScissorTl) TlY() uint32               { return Reg(r).Extract(16, 0x7FFF) } // TODO: Unused in pipeline
func (r PaScGenericScissorTl) WindowOffsetDisable() bool { return Reg(r).ExtractBool(31) }     // TODO: Unused in pipeline

type PaScWindowScissorTl Reg

func (r PaScWindowScissorTl) TlX() uint32               { return Reg(r).Extract(0, 0x7FFF) }  // TODO: Unused in pipeline
func (r PaScWindowScissorTl) TlY() uint32               { return Reg(r).Extract(16, 0x7FFF) } // TODO: Unused in pipeline
func (r PaScWindowScissorTl) WindowOffsetDisable() bool { return Reg(r).ExtractBool(31) }     // TODO: Unused in pipeline

type PaScWindowOffset Reg

func (r PaScWindowOffset) WindowXOffset() uint32 { return Reg(r).Extract(0, 0xFFFF) }
func (r PaScWindowOffset) WindowYOffset() uint32 { return Reg(r).Extract(16, 0xFFFF) }

type PaSuHardwareScreenOffset Reg

func (r PaSuHardwareScreenOffset) HwScreenOffsetX() uint32 { return Reg(r).Extract(0, 0x1FF) }  // TODO: Unused in pipeline
func (r PaSuHardwareScreenOffset) HwScreenOffsetY() uint32 { return Reg(r).Extract(16, 0x1FF) } // TODO: Unused in pipeline

type PaClVsOutCntl uint32

func (r PaClVsOutCntl) ClipDistEna0() bool           { return Reg(r).ExtractBool(0) }
func (r PaClVsOutCntl) ClipDistEna1() bool           { return Reg(r).ExtractBool(1) }
func (r PaClVsOutCntl) ClipDistEna2() bool           { return Reg(r).ExtractBool(2) }
func (r PaClVsOutCntl) ClipDistEna3() bool           { return Reg(r).ExtractBool(3) }
func (r PaClVsOutCntl) ClipDistEna4() bool           { return Reg(r).ExtractBool(4) }
func (r PaClVsOutCntl) ClipDistEna5() bool           { return Reg(r).ExtractBool(5) }
func (r PaClVsOutCntl) ClipDistEna6() bool           { return Reg(r).ExtractBool(6) }
func (r PaClVsOutCntl) ClipDistEna7() bool           { return Reg(r).ExtractBool(7) }
func (r PaClVsOutCntl) CullDistEna0() bool           { return Reg(r).ExtractBool(8) }
func (r PaClVsOutCntl) CullDistEna1() bool           { return Reg(r).ExtractBool(9) }
func (r PaClVsOutCntl) CullDistEna2() bool           { return Reg(r).ExtractBool(10) }
func (r PaClVsOutCntl) CullDistEna3() bool           { return Reg(r).ExtractBool(11) }
func (r PaClVsOutCntl) CullDistEna4() bool           { return Reg(r).ExtractBool(12) }
func (r PaClVsOutCntl) CullDistEna5() bool           { return Reg(r).ExtractBool(13) }
func (r PaClVsOutCntl) CullDistEna6() bool           { return Reg(r).ExtractBool(14) }
func (r PaClVsOutCntl) CullDistEna7() bool           { return Reg(r).ExtractBool(15) }
func (r PaClVsOutCntl) UseVtxPointSize() bool        { return Reg(r).ExtractBool(16) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) UseVtxEdgeFlag() bool         { return Reg(r).ExtractBool(17) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) UseVtxRenderTargetIndx() bool { return Reg(r).ExtractBool(18) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) UseVtxViewportIndx() bool     { return Reg(r).ExtractBool(19) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) UseVtxKillFlag() bool         { return Reg(r).ExtractBool(20) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) VsOutMiscVecEna() bool        { return Reg(r).ExtractBool(21) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) VsOutCcdist0VecEna() bool     { return Reg(r).ExtractBool(22) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) ClipDistEna() uint8           { return uint8(Reg(r).Extract(0, 0xFF)) }
func (r PaClVsOutCntl) CullDistEna() uint8           { return uint8(Reg(r).Extract(8, 0xFF)) }
func (r PaClVsOutCntl) VsOutMiscSideBusEna() bool    { return Reg(r).ExtractBool(24) } // TODO: Unused in pipeline
func (r PaClVsOutCntl) UseVtxGsCutFlag() bool        { return Reg(r).ExtractBool(25) } // TODO: Unused in pipeline

type PaSuLineCntl Reg

func (r PaSuLineCntl) Width() uint32 { return Reg(r).Extract(0, 0xFFFF) }

type PaScAaMaskX0y0X1y0 Reg

func (r PaScAaMaskX0y0X1y0) AaMaskX0y0() uint32 { return Reg(r).Extract(0, 0xFFFF) }
func (r PaScAaMaskX0y0X1y0) AaMaskX1y0() uint32 { return Reg(r).Extract(16, 0xFFFF) }

type PaScAaMaskX0y1X1y1 Reg

func (r PaScAaMaskX0y1X1y1) AaMaskX0y1() uint32 { return Reg(r).Extract(0, 0xFFFF) }
func (r PaScAaMaskX0y1X1y1) AaMaskX1y1() uint32 { return Reg(r).Extract(16, 0xFFFF) }

type PaSuPolyOffsetClamp Reg

func (r PaSuPolyOffsetClamp) Clamp() uint32 { return Reg(r).Extract(0, 0xFFFFFFFF) }

type PaSuPolyOffsetFrontScale Reg

func (r PaSuPolyOffsetFrontScale) Scale() uint32 { return Reg(r).Extract(0, 0xFFFFFFFF) }

type PaSuPolyOffsetFrontOffset Reg

func (r PaSuPolyOffsetFrontOffset) Offset() uint32 { return Reg(r).Extract(0, 0xFFFFFFFF) }

type PaSuPolyOffsetBackScale Reg

func (r PaSuPolyOffsetBackScale) Scale() uint32 { return Reg(r).Extract(0, 0xFFFFFFFF) }

type PaSuPolyOffsetBackOffset Reg

func (r PaSuPolyOffsetBackOffset) Offset() uint32 { return Reg(r).Extract(0, 0xFFFFFFFF) }

type PaSuVtxCntl Reg

func (r PaSuVtxCntl) PixCenter() uint32 { return Reg(r).Extract(0, 0x1) } // TODO: Unused in pipeline
func (r PaSuVtxCntl) RoundMode() uint32 { return Reg(r).Extract(1, 0x3) } // TODO: Unused in pipeline
func (r PaSuVtxCntl) QuantMode() uint32 { return Reg(r).Extract(3, 0x7) } // TODO: Unused in pipeline

type PaScCliprectRule Reg

func (r PaScCliprectRule) ClipRule() uint32 { return Reg(r).Extract(0, 0xFFFF) } // TODO: Unused in pipeline
