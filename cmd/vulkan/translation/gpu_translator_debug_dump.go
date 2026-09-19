package translation

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

func PrimTypeName(primType uint32) string {
	switch primType {
	case 1:
		return "POINTLIST"
	case 2:
		return "LINELIST"
	case 3:
		return "LINESTRIP"
	case 4:
		return "TRILIST"
	case 5:
		return "TRIFAN"
	case 6:
		return "TRISTRIP"
	case 9:
		return "PATCH"
	case 10:
		return "LINELIST_ADJ"
	case 11:
		return "LINESTRIP_ADJ"
	case 12:
		return "TRILIST_ADJ"
	case 13:
		return "TRISTRIP_ADJ"
	case 16:
		return "TRI_WITH_WFLAGS"
	case 17:
		return "RECTLIST"
	case 18:
		return "LINELOOP"
	case 19:
		return "QUADLIST"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", primType)
	}
}

func (t *GpuTranslator) EnableDebugDump(frame uint64) {
	t.debugDumpMutex.Lock()
	defer t.debugDumpMutex.Unlock()
	t.debugDumpEnabled = true
	t.debugDumpFrame = frame
	t.debugDumpLog.Reset()
	t.debugDumpLog.WriteString(fmt.Sprintf("==================== FRAME %d DEBUG DUMP ====================\n", frame))
}

func (t *GpuTranslator) DisableDebugDump() {
	t.debugDumpMutex.Lock()
	defer t.debugDumpMutex.Unlock()
	t.debugDumpEnabled = false
}

func (t *GpuTranslator) IsDebugDumpEnabled() bool {
	t.debugDumpMutex.Lock()
	defer t.debugDumpMutex.Unlock()
	return t.debugDumpEnabled
}

func (t *GpuTranslator) LogDebug(format string, args ...any) {
	t.debugDumpMutex.Lock()
	defer t.debugDumpMutex.Unlock()
	if t.debugDumpEnabled {
		t.debugDumpLog.WriteString(fmt.Sprintf(format, args...))
		t.debugDumpLog.WriteByte('\n')
	}
}

func (t *GpuTranslator) GetAndClearDebugDump() (string, uint64) {
	t.debugDumpMutex.Lock()
	defer t.debugDumpMutex.Unlock()
	t.debugDumpEnabled = false
	frame := t.debugDumpFrame
	res := t.debugDumpLog.String()
	t.debugDumpLog.Reset()
	return res, frame
}

func CompareOpName(op uint32) string {
	switch op {
	case 0:
		return "NEVER"
	case 1:
		return "LESS"
	case 2:
		return "EQUAL"
	case 3:
		return "LEQUAL"
	case 4:
		return "GREATER"
	case 5:
		return "NOTEQUAL"
	case 6:
		return "GEQUAL"
	case 7:
		return "ALWAYS"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", op)
	}
}

func (t *GpuTranslator) RecordBindPipelineDebug(bind *gpu.LiverpoolBindPipeline) {
	if !t.IsDebugDumpEnabled() {
		return
	}
	t.LogDebug("--- BIND PIPELINE ---")
	t.LogDebug("  Color Target: Addr=0x%016X, Size=%dx%d, Pitch=%d, Format=%d, NumType=%d, CompSwap=%d",
		bind.RtBase.Address(), bind.RtWidth, bind.RtHeight, bind.RtPitch,
		bind.CbColorInfo0.Format(), bind.CbColorInfo0.NumberType(), bind.CbColorInfo0.CompSwap())
	t.LogDebug("  Depth Target: Addr=0x%016X, Size=%dx%d, Format=%d, ZEnable=%t, ZWriteEnable=%t, ZFunc=%d (%s), ClearVal=0x%08X (%f), ClearEna=%t",
		bind.DbZWriteBase.Address(), bind.DbWidth, bind.DbHeight,
		bind.DbZInfo.Format(), bind.DbDepthControl.ZEnable(), bind.DbDepthControl.ZWriteEnable(),
		bind.DbDepthControl.Zfunc(), CompareOpName(bind.DbDepthControl.Zfunc()),
		bind.DbDepthClearValue, math.Float32frombits(bind.DbDepthClearValue),
		bind.DbRenderControl.DepthClearEnable())
	t.LogDebug("  Topology/PrimType: %d (%s)", bind.PrimType, PrimTypeName(bind.PrimType))
	t.LogDebug("  Blend Control: Enable=%t, ROP3Disable=%t, ColorMask=0x%X",
		bind.RtBlendControl.Enable(), bind.RtBlendControl.DisableRop3(), bind.RtTargetMask)
}

func (t *GpuTranslator) RecordDynamicStateDebug(dyn *gpu.LiverpoolSetDynamicState) {
	if !t.IsDebugDumpEnabled() {
		return
	}
	t.LogDebug("--- SET DYNAMIC STATE ---")
	t.LogDebug("  Viewport Registers: Scale=(%f, %f, %f) Offset=(%f, %f, %f)",
		dyn.VpXScale, dyn.VpYScale, dyn.VpZScale,
		dyn.VpXOffset, dyn.VpYOffset, dyn.VpZOffset)
	t.LogDebug("  PaClVteCntl (0x%08X): XScaleEna=%t, XOffsetEna=%t, YScaleEna=%t, YOffsetEna=%t, ZScaleEna=%t, ZOffsetEna=%t",
		uint32(dyn.PaClVteCntl),
		dyn.PaClVteCntl.VpXScaleEnable(), dyn.PaClVteCntl.VpXOffsetEnable(),
		dyn.PaClVteCntl.VpYScaleEnable(), dyn.PaClVteCntl.VpYOffsetEnable(),
		dyn.PaClVteCntl.VpZScaleEnable(), dyn.PaClVteCntl.VpZOffsetEnable())
	t.LogDebug("  PaClClipCntl (0x%08X): ClipDisable=%t, DxClipSpaceDef=%t",
		uint32(dyn.ClipControl), dyn.ClipControl.ClipDisable(), dyn.ClipControl.DxClipSpaceDef())
	t.LogDebug("  Scissors: Screen=[%d,%d..%d,%d] Window=[%d,%d..%d,%d] Vp=[%d,%d..%d,%d] Generic=[%d,%d..%d,%d]",
		dyn.ScissorTl.TlX(), dyn.ScissorTl.TlY(), dyn.ScissorBr&0xFFFF, (dyn.ScissorBr>>16)&0xFFFF,
		dyn.WindowScissorTl&0x7FFF, (dyn.WindowScissorTl>>16)&0x7FFF, dyn.WindowScissorBr&0x7FFF, (dyn.WindowScissorBr>>16)&0x7FFF,
		dyn.VpScissorTl&0x7FFF, (dyn.VpScissorTl>>16)&0x7FFF, dyn.VpScissorBr&0x7FFF, (dyn.VpScissorBr>>16)&0x7FFF,
		dyn.GenericScissorTl&0x7FFF, (dyn.GenericScissorTl>>16)&0x7FFF, dyn.GenericScissorBr&0x7FFF, (dyn.GenericScissorBr>>16)&0x7FFF)
}

func (t *GpuTranslator) RecordDrawDebug(frame uint64, draw *gpu.LiverpoolDraw, pushVs spirvStructs.PushConstants) {
	if !t.IsDebugDumpEnabled() {
		return
	}
	t.LogDebug("--- DRAW CALL ---")
	t.LogDebug("  Indices: Count=%d, InstanceCount=%d, PrimType=%d (%s), Indexed=%t, IndexOffset=%d, IndexBase=0x%016X",
		draw.IndexCount, draw.InstanceCount, draw.PrimType, PrimTypeName(draw.PrimType), draw.IsIndexed, draw.IndexOffset, draw.IndexBase)
	t.LogDebug("  Shaders: VS=0x%016X, FS=0x%016X",
		t.activeVertexShaderKey.Address, t.activeFragmentShaderKey.Address)
	t.LogDebug("  PushConstants: VpXScale=%f, VpXOffset=%f, VpYScale=%f, VpYOffset=%f, ClipControl=0x%08X (ClipDisable=%t)",
		pushVs.VpXScale, pushVs.VpXOffset, pushVs.VpYScale, pushVs.VpYOffset,
		pushVs.ClipControl, (pushVs.ClipControl&(1<<16)) != 0)
	if draw.DbRenderControl.DepthClearEnable() || draw.DbRenderControl.StencilClearEnable() {
		t.LogDebug("  Draw Clear: DepthClear=%t, StencilClear=%t, ClearVal=0x%08X (%f)",
			draw.DbRenderControl.DepthClearEnable(), draw.DbRenderControl.StencilClearEnable(),
			draw.DbDepthClearValue, math.Float32frombits(draw.DbDepthClearValue))
	}
}
