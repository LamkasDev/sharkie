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
	b.EmitCapability(SpvCapSubgroupBallotKHR)
	b.EmitCapability(SpvCapPhysicalStorageBufferAddresses)
	b.EmitExtension("SPV_KHR_physical_storage_buffer")
	idGLSL := b.EmitExtInstImport("GLSL.std.450")
	b.EmitMemoryModel(SpvAddrModelPhysicalStorageBuffer64, SpvMemModelGLSL450)

	// Common types.
	idVoid := b.EmitTypeVoid()
	idBool := b.EmitTypeBool()
	idInt := b.EmitTypeInt(32, true)
	idUint := b.EmitTypeInt(32, false)
	idUint64 := b.EmitTypeInt(64, false)
	idInt64 := b.EmitTypeInt(64, true)
	idFnType := b.EmitTypeFunction(idVoid)

	idFloat := b.EmitTypeFloat(32)
	idV2Float := b.EmitTypeVector(idFloat, 2)
	idV4Float := b.EmitTypeVector(idFloat, 4)
	idV4Uint := b.EmitTypeVector(idUint, 4)

	idTrue := b.EmitConstantTrue(idBool)
	idFalse := b.EmitConstantFalse(idBool)

	// Built-ins.
	idPtrInputUint := b.EmitTypePointer(SpvStorageInput, idUint)
	idSubgroupLocalInvocationId := b.EmitVariable(idPtrInputUint, SpvStorageInput)
	b.EmitDecorate(idSubgroupLocalInvocationId, SpvDecorationBuiltIn, SpvBuiltInSubgroupLocalInvocationId)

	// Types for constant RAM access via PhysicalStorageBuffer.
	idPtrPsbUint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, idUint)

	// Push constants.
	// struct StubPushConstants {
	// 	float Time;
	// 	uint32 _;
	// 	PhysicalStorageBuffer uint* ConstRamAddress;
	// 	PhysicalStorageBuffer uint* UserDataAddress;
	// 	uint64 GarlicAddress;
	// 	uint64 OnionAddress;
	// }
	idUd := b.EmitTypeStruct(idFloat, idUint, idPtrPsbUint, idPtrPsbUint, idUint64, idUint64)
	idPtrPc := b.EmitTypePointer(SpvStoragePushConstant, idUd)
	idPtrPcFloat := b.EmitTypePointer(SpvStoragePushConstant, idFloat)
	idPtrPcPsbUint := b.EmitTypePointer(SpvStoragePushConstant, idPtrPsbUint)
	idPtrPcPsbUint64 := b.EmitTypePointer(SpvStoragePushConstant, idUint64)

	// Annotations for the push-constant block.
	b.EmitDecorate(idUd, SpvDecorationBlock)
	b.EmitMemberDecorate(idUd, 0, SpvDecorationOffset, 0)
	b.EmitMemberDecorate(idUd, 1, SpvDecorationOffset, 4)
	b.EmitMemberDecorate(idUd, 2, SpvDecorationOffset, 8)
	b.EmitMemberDecorate(idUd, 3, SpvDecorationOffset, 16)
	b.EmitMemberDecorate(idUd, 4, SpvDecorationOffset, 24)
	b.EmitMemberDecorate(idUd, 5, SpvDecorationOffset, 32)

	// Global push-constant variable.
	idPCVar := b.EmitVariable(idPtrPc, SpvStoragePushConstant)
	b.EmitDecorate(idPCVar, SpvDecorationAliasedPointer)

	// Stage-specific outputs.
	interfaceIds := []uint32{idSubgroupLocalInvocationId}
	var idPosOut, idFragDepthOut uint32
	var idColorOuts [8]uint32
	var idParamOuts [32]uint32
	var idZeroVec4 uint32

	idPtrOutV4 := b.EmitTypePointer(SpvStorageOutput, idV4Float)
	idPtrOutF := b.EmitTypePointer(SpvStorageOutput, idFloat)

	if shader.Stage == GcnShaderStageVertex {
		idPosOut = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
		b.EmitDecorate(idPosOut, SpvDecorationBuiltIn, SpvBuiltInPosition)
		interfaceIds = append(interfaceIds, idPosOut)

		for i := range idParamOuts {
			idParamOuts[i] = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
			b.EmitDecorate(idParamOuts[i], SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idParamOuts[i])
		}
	} else if shader.Stage == GcnShaderStageFragment {
		for i := range idColorOuts {
			idColorOuts[i] = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
			b.EmitDecorate(idColorOuts[i], SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idColorOuts[i])
		}

		idFragDepthOut = b.EmitVariable(idPtrOutF, SpvStorageOutput)
		b.EmitDecorate(idFragDepthOut, SpvDecorationBuiltIn, SpvBuiltInFragDepth)
		interfaceIds = append(interfaceIds, idFragDepthOut)

		// Constant zero vec4 written on exit.
		idZeroF := b.EmitConstantFloat(idFloat, 0.0)
		idOneF := b.EmitConstantFloat(idFloat, 1.0)
		idZeroVec4 = b.EmitConstantComposite(idV4Float, idZeroF, idZeroF, idZeroF, idOneF)
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

	// Register GCN SGPRs and VGPRs.
	idPtrFnUint := b.EmitTypePointer(SpvStorageFunction, idUint)
	var gcnSgprIds [104]uint32
	for i := range gcnSgprIds {
		gcnSgprIds[i] = b.AllocId()
	}
	var gcnVgprIds [256]uint32
	for i := range gcnVgprIds {
		gcnVgprIds[i] = b.AllocId()
	}

	// GCN special registers.
	var gcnSpecialIds [27]uint32
	gcnSpecialIds[GcnSpecIdxFlatScrLo] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxFlatScrHi] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxVccLo] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxVccHi] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxTbaLo] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxTbaHi] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxTmaLo] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxTmaHi] = b.AllocId()
	for i := range 12 {
		gcnSpecialIds[GcnSpecIdxTtmp0+i] = b.AllocId()
	}
	gcnSpecialIds[GcnSpecIdxM0] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxExecLo] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxExecHi] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxVccz] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxExecz] = b.AllocId()
	gcnSpecialIds[GcnSpecIdxScc] = b.AllocId()

	// GCN inline constants.
	var gcnConstIds [120]uint32
	gcnConstIds[GcnConstIdx0] = b.EmitConstantUint(idUint, 0)
	for i := uint32(GcnConstIdxInt1); i <= GcnConstIdxInt64; i++ {
		gcnConstIds[i] = b.EmitConstantUint(idUint, i)
	}
	for i := uint32(GcnConstIdxIntNeg1); i <= GcnConstIdxIntNeg16; i++ {
		val := uint32(int32(-(int(i) - GcnConstIdxInt64)))
		gcnConstIds[i] = b.EmitConstantUint(idUint, val)
	}
	gcnConstIds[GcnConstIdxFloat05] = b.EmitConstantUint(idUint, math.Float32bits(0.5))
	gcnConstIds[GcnConstIdxFloatNeg05] = b.EmitConstantUint(idUint, math.Float32bits(-0.5))
	gcnConstIds[GcnConstIdxFloat10] = b.EmitConstantUint(idUint, math.Float32bits(1.0))
	gcnConstIds[GcnConstIdxFloatNeg10] = b.EmitConstantUint(idUint, math.Float32bits(-1.0))
	gcnConstIds[GcnConstIdxFloat20] = b.EmitConstantUint(idUint, math.Float32bits(2.0))
	gcnConstIds[GcnConstIdxFloatNeg20] = b.EmitConstantUint(idUint, math.Float32bits(-2.0))
	gcnConstIds[GcnConstIdxFloat40] = b.EmitConstantUint(idUint, math.Float32bits(4.0))
	gcnConstIds[GcnConstIdxFloatNeg40] = b.EmitConstantUint(idUint, math.Float32bits(-4.0))

	// Function body.
	b.EmitFunction(idVoid, SpvFunctionControlNone, idFnType, idMain)

	// Emit reachable blocks in reverse post-order (entry block first).
	rpoBlockIds := shader.Cfg.ReversePostOrder()
	emittedBlockIds := make([]bool, len(shader.Cfg.Blocks))

	// Prepare internal IDs.
	ids := map[BlockContextId]uint32{
		BlockContextIdFalse:                     idFalse,
		BlockContextIdTrue:                      idTrue,
		BlockContextIdTypeBool:                  idBool,
		BlockContextIdTypeFloat:                 idFloat,
		BlockContextIdTypeInt:                   idInt,
		BlockContextIdTypeUint:                  idUint,
		BlockContextIdTypeUint64:                idUint64,
		BlockContextIdTypeInt64:                 idInt64,
		BlockContextIdTypeV2Float:               idV2Float,
		BlockContextIdTypeV4Float:               idV4Float,
		BlockContextIdTypeV4Uint:                idV4Uint,
		BlockContextIdPtrPcFloat:                idPtrPcFloat,
		BlockContextIdPtrPcPsbUint:              idPtrPcPsbUint,
		BlockContextIdPtrPcPsbUint64:            idPtrPcPsbUint64,
		BlockContextIdPtrPsbUint:                idPtrPsbUint,
		BlockContextIdPtrFnUint:                 idPtrFnUint,
		BlockContextIdPosOut:                    idPosOut,
		BlockContextIdFragDepthOut:              idFragDepthOut,
		BlockContextIdZeroVec4:                  idZeroVec4,
		BlockContextIdPcVar:                     idPCVar,
		BlockContextIdGlsl:                      idGLSL,
		BlockContextIdSubgroupLocalInvocationId: idSubgroupLocalInvocationId,
	}
	for i, id := range idColorOuts {
		ids[BlockContextIdColorOut0+BlockContextId(i)] = id
	}
	for i, id := range idParamOuts {
		ids[BlockContextIdParamOut0+BlockContextId(i)] = id
	}

	// Prepare block context with all GCN and our internal IDs.
	blockContext := SpirvBlockContext{
		Stage:    shader.Stage,
		LabelIds: labelIds,
		Ids:      ids,
		ConstIds: map[BlockContextId]uint32{
			ConstIdxUint0:        b.EmitConstantUint(idUint, 0),
			ConstIdxUint1:        b.EmitConstantUint(idUint, 1),
			ConstIdxUint2:        b.EmitConstantUint(idUint, 2),
			ConstIdxUint3:        b.EmitConstantUint(idUint, 3),
			ConstIdxUint4:        b.EmitConstantUint(idUint, 4),
			ConstIdxUint5:        b.EmitConstantUint(idUint, 5),
			ConstIdxUint6:        b.EmitConstantUint(idUint, 6),
			ConstIdxUint7:        b.EmitConstantUint(idUint, 7),
			ConstIdxUint32:       b.EmitConstantUint(idUint, 32),
			ConstIdxUint63:       b.EmitConstantUint(idUint, 63),
			ConstIdxUintFFFF:     b.EmitConstantUint(idUint, 0xFFFF),
			ConstIdxUint11111111: b.EmitConstantUint(idUint, 0x11111111),
			ConstIdxUintFFFFFFFF: b.EmitConstantUint(idUint, 0xFFFFFFFF),
			ConstIdxFloat1:       b.EmitConstantFloat(idFloat, 1.0),
			ConstIdxFloat0:       b.EmitConstantFloat(idFloat, 0.0),
		},
		GcnSgprIds:    gcnSgprIds,
		GcnVgprIds:    gcnVgprIds,
		GcnSpecialIds: gcnSpecialIds,
		GcnConstIds:   gcnConstIds,
	}

	// Emit reachable blocks.
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
