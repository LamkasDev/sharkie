package spirv

import (
	"math"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type SpirvShaderContext struct {
	NumThreads [3]uint32
}

type SpirvShader struct {
	Stage   GcnShaderStage
	Address uintptr
	Code    []uint32
}

func NewSpirvShader(shader *GcnShader, ctx SpirvShaderContext) (*SpirvShader, error) {
	b := NewSpvBuilder()

	// Capabilities.
	b.EmitCapability(SpvCapShader)
	b.EmitCapability(SpvCapAddresses)
	b.EmitCapability(SpvCapInt64)
	b.EmitCapability(SpvCapPhysicalStorageBufferAddresses)
	b.EmitExtension("SPV_KHR_physical_storage_buffer")
	idGLSL := b.EmitExtInstImport("GLSL.std.450")
	b.EmitMemoryModel(SpvAddrModelPhysicalStorageBuffer64, SpvMemModelGLSL450)

	// Common types.
	idVoid := b.EmitTypeVoid()
	idBool := b.EmitTypeBool()
	idUint := b.EmitTypeInt(32, false)
	idUint64 := b.EmitTypeInt(64, false)
	idFnType := b.EmitTypeFunction(idVoid)

	idFloat := b.EmitTypeFloat(32)
	idV2Float := b.EmitTypeVector(idFloat, 2)
	idV4Float := b.EmitTypeVector(idFloat, 4)

	// Types for constant RAM access via PhysicalStorageBuffer.
	idPtrPsbUint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, idUint)

	// Push constants.
	// struct StubPushConstants {
	// 	float Time;
	// 	uint32 _;
	// 	PhysicalStorageBuffer uint* ConstRamAddress;
	// 	PhysicalStorageBuffer uint* UserDataAddress;
	// }
	idUd := b.EmitTypeStruct(idFloat, idUint, idPtrPsbUint, idPtrPsbUint)
	idPtrPc := b.EmitTypePointer(SpvStoragePushConstant, idUd)
	idPtrPcFloat := b.EmitTypePointer(SpvStoragePushConstant, idFloat)
	idPtrPcPsbUint := b.EmitTypePointer(SpvStoragePushConstant, idPtrPsbUint)

	// Annotations for the push-constant block.
	b.EmitDecorate(idUd, SpvDecorationBlock)
	b.EmitMemberDecorate(idUd, 0, SpvDecorationOffset, 0)
	b.EmitMemberDecorate(idUd, 1, SpvDecorationOffset, 4)
	b.EmitMemberDecorate(idUd, 2, SpvDecorationOffset, 8)
	b.EmitMemberDecorate(idUd, 3, SpvDecorationOffset, 16)

	// Global push-constant variable.
	idPCVar := b.EmitVariable(idPtrPc, SpvStoragePushConstant)
	b.EmitDecorate(idPCVar, SpvDecorationAliasedPointer)

	// Stub boolean constant (condition in untranslated conditional branches).
	idFalse := b.EmitConstantFalse(idBool)

	// Stage-specific outputs.
	var interfaceIds []uint32
	var idColorOut, idZeroVec4 uint32
	if shader.Stage == GcnShaderStageFragment {
		// Declare vec4 color output at location 0.
		idPtrOutV4 := b.EmitTypePointer(SpvStorageOutput, idV4Float)
		idColorOut = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
		b.EmitDecorate(idColorOut, SpvDecorationLocation, 0)

		// Constant zero vec4 written on exit.
		idZeroF := b.EmitConstantFloat(idFloat, 0.0)
		idOneF := b.EmitConstantFloat(idFloat, 1.0)
		idZeroVec4 = b.EmitConstantComposite(idV4Float, idZeroF, idZeroF, idZeroF, idOneF)

		interfaceIds = append(interfaceIds, idColorOut)
	}

	// Entry point.
	idMain := b.AllocId()
	b.EmitEntryPoint(GncStageToSpvExecModel[shader.Stage], idMain, "main", interfaceIds...)

	// Execution modes.
	switch shader.Stage {
	case GcnShaderStageFragment:
		b.EmitExecutionMode(idMain, SpvExecModeOriginUpperLeft)
	case GcnShaderStageCompute:
		numThreads := ctx.NumThreads
		for i := range numThreads {
			if numThreads[i] == 0 {
				numThreads[i] = 1
			}
		}
		b.EmitExecutionMode(idMain, SpvExecModeLocalSize, numThreads[0], numThreads[1], numThreads[2])
	}

	// Pre-allocate SPIR-V labels ID for GCN CFG blocks.
	labelIds := make([]uint32, len(shader.Cfg.Blocks))
	for i := range shader.Cfg.Blocks {
		labelIds[i] = b.AllocId()
	}

	// Register SGPRs and VGPRs.
	idPtrFnUint := b.EmitTypePointer(SpvStorageFunction, idUint)
	var sgprIds [104]uint32
	for i := range sgprIds {
		sgprIds[i] = b.AllocId()
	}
	var vgprIds [256]uint32
	for i := range vgprIds {
		vgprIds[i] = b.AllocId()
	}

	// Special registers.
	var specialIds [25]uint32
	specialIds[SpecIdxVccLo] = b.AllocId()
	specialIds[SpecIdxVccHi] = b.AllocId()
	specialIds[SpecIdxTbaLo] = b.AllocId()
	specialIds[SpecIdxTbaHi] = b.AllocId()
	specialIds[SpecIdxTmaLo] = b.AllocId()
	specialIds[SpecIdxTmaHi] = b.AllocId()
	for i := range 12 {
		specialIds[SpecIdxTtmp0+i] = b.AllocId()
	}
	specialIds[SpecIdxM0] = b.AllocId()
	specialIds[SpecIdxExecLo] = b.AllocId()
	specialIds[SpecIdxExecHi] = b.AllocId()
	specialIds[SpecIdxVccz] = b.AllocId()
	specialIds[SpecIdxExecz] = b.AllocId()
	specialIds[SpecIdxScc] = b.AllocId()

	// Inline constants.
	var constIds [120]uint32
	constIds[ConstIdx0] = b.EmitConstantUint(idUint, 0)
	for i := uint32(ConstIdxInt1); i <= ConstIdxInt64; i++ {
		constIds[i] = b.EmitConstantUint(idUint, i)
	}
	for i := uint32(ConstIdxIntNeg1); i <= ConstIdxIntNeg16; i++ {
		val := uint32(int32(-(int(i) - ConstIdxInt64)))
		constIds[i] = b.EmitConstantUint(idUint, val)
	}
	constIds[ConstIdxFloat05] = b.EmitConstantUint(idUint, math.Float32bits(0.5))
	constIds[ConstIdxFloatNeg05] = b.EmitConstantUint(idUint, math.Float32bits(-0.5))
	constIds[ConstIdxFloat10] = b.EmitConstantUint(idUint, math.Float32bits(1.0))
	constIds[ConstIdxFloatNeg10] = b.EmitConstantUint(idUint, math.Float32bits(-1.0))
	constIds[ConstIdxFloat20] = b.EmitConstantUint(idUint, math.Float32bits(2.0))
	constIds[ConstIdxFloatNeg20] = b.EmitConstantUint(idUint, math.Float32bits(-2.0))
	constIds[ConstIdxFloat40] = b.EmitConstantUint(idUint, math.Float32bits(4.0))
	constIds[ConstIdxFloatNeg40] = b.EmitConstantUint(idUint, math.Float32bits(-4.0))

	// Function body.
	b.EmitFunction(idVoid, SpvFunctionControlNone, idFnType, idMain)

	// Emit reachable blocks in reverse post-order (entry block first).
	rpoBlockIds := shader.Cfg.ReversePostOrder()
	emittedBlockIds := make([]bool, len(shader.Cfg.Blocks))
	blockContext := SpirvBlockContext{
		Stage:    shader.Stage,
		LabelIds: labelIds,
		Ids: map[SpirvBlockContextId]uint32{
			SpirvBlockContextIdFalse:        idFalse,
			SpirvBlockContextIdColorOut:     idColorOut,
			SpirvBlockContextIdZeroVec4:     idZeroVec4,
			SpirvBlockContextIdPcVar:        idPCVar,
			SpirvBlockContextIdPtrPcFloat:   idPtrPcFloat,
			SpirvBlockContextIdPtrPcPsbUint: idPtrPcPsbUint,
			SpirvBlockContextIdPtrPsbUint:   idPtrPsbUint,
			SpirvBlockContextIdUint:         idUint,
			SpirvBlockContextIdUint64:       idUint64,
			SpirvBlockContextIdFloat:        idFloat,
			SpirvBlockContextIdV2Float:      idV2Float,
			SpirvBlockContextIdV4Float:      idV4Float,
			SpirvBlockContextIdPtrFnUint:    idPtrFnUint,
			SpirvBlockContextIdGlsl:         idGLSL,
		},
		SgprIds:    sgprIds,
		VgprIds:    vgprIds,
		SpecialIds: specialIds,
		ConstIds:   constIds,
	}
	for _, blockId := range rpoBlockIds {
		emitBlock(b, &shader.Cfg.Blocks[blockId], blockContext)
		emittedBlockIds[blockId] = true
	}

	// Emit any unreachable blocks.
	for i := range shader.Cfg.Blocks {
		if !emittedBlockIds[i] {
			b.EmitLabel(labelIds[i])
			b.EmitUnreachable()
		}
	}

	// Andddd we're done :)
	b.EmitFunctionEnd()

	return &SpirvShader{
		Address: shader.Address,
		Stage:   shader.Stage,
		Code:    b.Assemble(),
	}, nil
}
