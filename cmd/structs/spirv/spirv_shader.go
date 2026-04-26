package spirv

import (
	"fmt"
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
	b.EmitCapability(SpvCapSampled1D)
	b.EmitCapability(SpvCapImageQuery)
	b.EmitCapability(SpvCapGroupNonUniformBallot)
	b.EmitCapability(SpvCapSubgroupBallotKHR)
	b.EmitCapability(SpvCapRuntimeDescriptorArray)
	b.EmitCapability(SpvCapPhysicalStorageBufferAddresses)
	b.EmitExtension("SPV_KHR_physical_storage_buffer")
	b.EmitExtension("SPV_KHR_shader_ballot")
	b.EmitExtension("SPV_EXT_descriptor_indexing")
	typeGLSL := b.EmitExtInstImport("GLSL.std.450")
	b.EmitMemoryModel(SpvAddrModelPhysicalStorageBuffer64, SpvMemModelGLSL450)

	// Common types.
	typeVoid := b.EmitTypeVoid()
	typeBool := b.EmitTypeBool()
	typeInt := b.EmitTypeInt(32, true)
	typeInt64 := b.EmitTypeInt(64, true)
	typeUint := b.EmitTypeInt(32, false)
	typeUint64 := b.EmitTypeInt(64, false)
	idFnType := b.EmitTypeFunction(typeVoid)

	typeV2Uint := b.EmitTypeVector(typeUint, 2)
	typeV3Uint := b.EmitTypeVector(typeUint, 3)
	typeV4Uint := b.EmitTypeVector(typeUint, 4)

	typeFloat := b.EmitTypeFloat(32)
	typeV2Float := b.EmitTypeVector(typeFloat, 2)
	typeV4Float := b.EmitTypeVector(typeFloat, 4)

	typeImage2d := b.EmitTypeImage(typeFloat, 1, 0, 0, 0, 1, 0)
	typeSampledImage2d := b.EmitTypeSampledImage(typeImage2d)
	typeBindlessArray2d := b.EmitTypeRuntimeArray(typeSampledImage2d)
	typePtrUniformSampledImage2d := b.EmitTypePointer(SpvStorageUniformConstant, typeSampledImage2d)
	typePtrUniformBindlessArray2d := b.EmitTypePointer(SpvStorageUniformConstant, typeBindlessArray2d)

	// Built-ins.
	idTrue := b.EmitConstantTrue(typeBool)
	idFalse := b.EmitConstantFalse(typeBool)

	typePtrInputUint := b.EmitTypePointer(SpvStorageInput, typeUint)
	typeSubgroupLocalInvocationId := b.EmitVariable(typePtrInputUint, SpvStorageInput)
	b.EmitDecorate(typeSubgroupLocalInvocationId, SpvDecorationBuiltIn, SpvBuiltInSubgroupLocalInvocationId)

	typeVertexIndex := b.EmitVariable(typePtrInputUint, SpvStorageInput)
	b.EmitName(typeVertexIndex, "vertex_index")
	b.EmitDecorate(typeVertexIndex, SpvDecorationBuiltIn, SpvBuiltInVertexIndex)

	typeInstanceIndex := b.EmitVariable(typePtrInputUint, SpvStorageInput)
	b.EmitName(typeInstanceIndex, "instance_index")
	b.EmitDecorate(typeInstanceIndex, SpvDecorationBuiltIn, SpvBuiltInInstanceIndex)

	// Push constants.
	// struct StubPushConstants {
	// 	PhysicalStorageBuffer uint* UserDataAddress;
	// }

	// Push constant types.
	typePtrPsbUint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, typeUint)
	typePtrPsbV2Uint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, typeV2Uint)
	typePtrPsbV3Uint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, typeV3Uint)
	typePtrPsbV4Uint := b.EmitTypePointer(SpvStoragePhysicalStorageBuffer, typeV4Uint)
	b.EmitDecorate(typePtrPsbUint, SpvDecorationArrayStride, 4)
	b.EmitDecorate(typePtrPsbV2Uint, SpvDecorationArrayStride, 8)
	b.EmitDecorate(typePtrPsbV3Uint, SpvDecorationArrayStride, 12)
	b.EmitDecorate(typePtrPsbV4Uint, SpvDecorationArrayStride, 16)

	// Push constant struct.
	idUd := b.EmitTypeStruct(typePtrPsbUint)
	idPtrPc := b.EmitTypePointer(SpvStoragePushConstant, idUd)
	typePtrPcPsbUint := b.EmitTypePointer(SpvStoragePushConstant, typePtrPsbUint)

	// Annotations for the push-constants.
	b.EmitDecorate(idUd, SpvDecorationBlock)
	b.EmitMemberDecorate(idUd, 0, SpvDecorationOffset, 0)

	// Global push-constant variable.
	typePcVar := b.EmitVariable(idPtrPc, SpvStoragePushConstant)
	b.EmitName(typePcVar, "push_constants")
	b.EmitDecorate(typePcVar, SpvDecorationAliasedPointer)

	// Bindless textures.
	typeBindlessTexturesVar := b.EmitVariable(typePtrUniformBindlessArray2d, SpvStorageUniformConstant)
	b.EmitName(typeBindlessTexturesVar, "bindless_textures")
	b.EmitDecorate(typeBindlessTexturesVar, SpvDecorationDescriptorSet, 0)
	b.EmitDecorate(typeBindlessTexturesVar, SpvDecorationBinding, 0)

	// Texel buffers (Set 1, Binding 0..3).
	var typeTexelBuffer uint32
	var idTexelBufferVars [4]uint32
	if shader.Stage == GcnShaderStageVertex {
		typeTexelBuffer = b.EmitTypeImage(typeFloat, 5, 0, 0, 0, 1, 0) // Dim=5 (Buffer)
		typePtrUniformTexelBuffer := b.EmitTypePointer(SpvStorageUniformConstant, typeTexelBuffer)
		for i := range 4 {
			idTexelBufferVars[i] = b.EmitVariable(typePtrUniformTexelBuffer, SpvStorageUniformConstant)
			b.EmitName(idTexelBufferVars[i], fmt.Sprintf("texel_buffer_%d", i))
			b.EmitDecorate(idTexelBufferVars[i], SpvDecorationDescriptorSet, 1)
			b.EmitDecorate(idTexelBufferVars[i], SpvDecorationBinding, uint32(i))
		}
	}

	// Stage-specific outputs.
	idPtrOutF := b.EmitTypePointer(SpvStorageOutput, typeFloat)
	idPtrOutV4 := b.EmitTypePointer(SpvStorageOutput, typeV4Float)

	interfaceIds := []uint32{typeSubgroupLocalInvocationId}
	var typePosOut, typeFragDepthOut uint32
	var idColorOuts [8]uint32
	var idParamOuts [32]uint32
	var typeZeroVec4 uint32
	switch shader.Stage {
	case GcnShaderStageVertex:
		typePosOut = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
		b.EmitName(typePosOut, "pos_out")
		b.EmitDecorate(typePosOut, SpvDecorationBuiltIn, SpvBuiltInPosition)

		interfaceIds = append(interfaceIds, typePosOut, typeVertexIndex, typeInstanceIndex)
		for i := range idParamOuts {
			idParamOuts[i] = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
			b.EmitDecorate(idParamOuts[i], SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idParamOuts[i])
		}
	case GcnShaderStageFragment:
		for i := range idColorOuts {
			idColorOuts[i] = b.EmitVariable(idPtrOutV4, SpvStorageOutput)
			b.EmitDecorate(idColorOuts[i], SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idColorOuts[i])
		}

		typeFragDepthOut = b.EmitVariable(idPtrOutF, SpvStorageOutput)
		b.EmitName(typeFragDepthOut, "frag_depth_out")
		b.EmitDecorate(typeFragDepthOut, SpvDecorationBuiltIn, SpvBuiltInFragDepth)
		interfaceIds = append(interfaceIds, typeFragDepthOut)

		// Constant zero vec4 written on exit.
		idZeroF := b.EmitConstantFloat(typeFloat, 0.0)
		idOneF := b.EmitConstantFloat(typeFloat, 1.0)
		typeZeroVec4 = b.EmitConstantComposite(typeV4Float, idZeroF, idZeroF, idZeroF, idOneF)
		b.EmitName(typeZeroVec4, "zero_vec4")
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

	// Register GCN SGPRs and VGPRs.
	typePtrFnUint := b.EmitTypePointer(SpvStorageFunction, typeUint)
	idSgprCount := b.EmitConstantUint(typeUint, 104)
	idVgprCount := b.EmitConstantUint(typeUint, 256)
	typeSgprArray := b.EmitTypeArray(typeUint, idSgprCount)
	typeVgprArray := b.EmitTypeArray(typeUint, idVgprCount)
	typePtrSgprArray := b.EmitTypePointer(SpvStorageFunction, typeSgprArray)
	typePtrVgprArray := b.EmitTypePointer(SpvStorageFunction, typeVgprArray)

	idSgprArrayVar := b.AllocId()
	idVgprArrayVar := b.AllocId()
	b.EmitName(idSgprArrayVar, "sgprs")
	b.EmitName(idVgprArrayVar, "vgprs")
	b.EmitDeferredLocalVariable(typePtrSgprArray, idSgprArrayVar)
	b.EmitDeferredLocalVariable(typePtrVgprArray, idVgprArrayVar)

	// GCN special registers.
	var gcnSpecialIds [27]SpirvBlockContextUsedId
	gcnSpecialIds[GcnSpecIdxFlatScrLo] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "flat_scr_lo"}
	gcnSpecialIds[GcnSpecIdxFlatScrHi] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "flat_scr_hi"}
	gcnSpecialIds[GcnSpecIdxVccLo] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "vcc_lo"}
	gcnSpecialIds[GcnSpecIdxVccHi] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "vcc_hi"}
	gcnSpecialIds[GcnSpecIdxTbaLo] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "tba_lo"}
	gcnSpecialIds[GcnSpecIdxTbaHi] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "tba_hi"}
	gcnSpecialIds[GcnSpecIdxTmaLo] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "tma_lo"}
	gcnSpecialIds[GcnSpecIdxTmaHi] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "tma_hi"}
	for i := range 12 {
		gcnSpecialIds[GcnSpecIdxTtmp0+i] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: fmt.Sprintf("ttmp%d", i)}
	}
	gcnSpecialIds[GcnSpecIdxM0] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "m0"}
	gcnSpecialIds[GcnSpecIdxExecLo] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "exec_lo"}
	gcnSpecialIds[GcnSpecIdxExecHi] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "exec_hi"}
	gcnSpecialIds[GcnSpecIdxVccz] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "vccz"}
	gcnSpecialIds[GcnSpecIdxExecz] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "execz"}
	gcnSpecialIds[GcnSpecIdxScc] = SpirvBlockContextUsedId{Id: b.AllocId(), Name: "scc"}

	// GCN inline constants.
	var gcnConstIds [120]SpirvBlockContextUsedId
	for i := range gcnConstIds {
		usedId := SpirvBlockContextUsedId{Id: b.AllocId()}
		switch {
		case i == GcnConstIdx0:
			usedId.Value, usedId.Name = 0, "0"
		case i >= GcnConstIdxInt1 && i <= GcnConstIdxInt64:
			usedId.Value, usedId.Name = uint32(i), fmt.Sprint(i)
		case i >= GcnConstIdxIntNeg1 && i <= GcnConstIdxIntNeg16:
			v := int32(-(int(i) - GcnConstIdxInt64))
			usedId.Value, usedId.Name = uint32(v), fmt.Sprint(v)
		case i == GcnConstIdxFloat05:
			usedId.Value, usedId.Name = math.Float32bits(0.5), "0.5"
		case i == GcnConstIdxFloatNeg05:
			usedId.Value, usedId.Name = math.Float32bits(-0.5), "-0.5"
		case i == GcnConstIdxFloat10:
			usedId.Value, usedId.Name = math.Float32bits(1.0), "1.0"
		case i == GcnConstIdxFloatNeg10:
			usedId.Value, usedId.Name = math.Float32bits(-1.0), "-1.0"
		case i == GcnConstIdxFloat20:
			usedId.Value, usedId.Name = math.Float32bits(2.0), "2.0"
		case i == GcnConstIdxFloatNeg20:
			usedId.Value, usedId.Name = math.Float32bits(-2.0), "-2.0"
		case i == GcnConstIdxFloat40:
			usedId.Value, usedId.Name = math.Float32bits(4.0), "4.0"
		case i == GcnConstIdxFloatNeg40:
			usedId.Value, usedId.Name = math.Float32bits(-4.0), "-4.0"
		}
		gcnConstIds[i] = usedId
	}

	// Internal inline constants.
	constIds := map[BlockContextId]SpirvBlockContextUsedId{}
	for i := range BlockContextId(256) {
		constIds[i] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: uint32(i), Name: fmt.Sprint(i)}
	}
	constIds[ConstIdxUint3FFF] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: 0x3FFF, Name: "0x3FFF"}
	constIds[ConstIdxUintFFFF] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: 0xFFFF, Name: "0xFFFF"}
	constIds[ConstIdxUint11111111] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: 0x11111111, Name: "0x11111111"}
	constIds[ConstIdxUintFFFFFFFF] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: 0xFFFFFFFF, Name: "0xFFFFFFFF"}
	constIds[ConstIdx64Uint0] = SpirvBlockContextUsedId{Id: b.AllocId(), Value64: 0, Name: "64_0"}
	constIds[ConstIdx64Uint32] = SpirvBlockContextUsedId{Id: b.AllocId(), Value64: 32, Name: "64_32"}
	constIds[ConstIdxFloat1] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: math.Float32bits(1.0), Name: "1.0"}
	constIds[ConstIdxFloat0] = SpirvBlockContextUsedId{Id: b.AllocId(), Value: math.Float32bits(0.0), Name: "0.0"}

	// Prepare internal IDs.
	ids := map[BlockContextId]SpirvBlockContextUsedId{
		BlockContextIdFalse:                     {Id: idFalse, Name: "false"},
		BlockContextIdTrue:                      {Id: idTrue, Name: "true"},
		BlockContextIdTypeBool:                  {Id: typeBool, Name: "bool_t"},
		BlockContextIdTypeFloat:                 {Id: typeFloat, Name: "float_t"},
		BlockContextIdTypeInt:                   {Id: typeInt, Name: "int_t"},
		BlockContextIdTypeUint:                  {Id: typeUint, Name: "uint_t"},
		BlockContextIdTypeUint64:                {Id: typeUint64, Name: "uint64_t"},
		BlockContextIdTypeInt64:                 {Id: typeInt64, Name: "int64_t"},
		BlockContextIdTypeV2Float:               {Id: typeV2Float, Name: "v2float_t"},
		BlockContextIdTypeV4Float:               {Id: typeV4Float, Name: "v4float_t"},
		BlockContextIdTypeV2Uint:                {Id: typeV2Uint, Name: "v2uint_t"},
		BlockContextIdTypeV3Uint:                {Id: typeV3Uint, Name: "v3uint_t"},
		BlockContextIdTypeV4Uint:                {Id: typeV4Uint, Name: "v4uint_t"},
		BlockContextIdTypeSampledImage:          {Id: typeSampledImage2d, Name: "sampled_image_2d_t"},
		BlockContextIdPtrUniformSampledImage:    {Id: typePtrUniformSampledImage2d, Name: "ptr_uniform_sampled_image_2d_t"},
		BlockContextIdPtrPcPsbUint:              {Id: typePtrPcPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPsbUint:                {Id: typePtrPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPsbV2Uint:              {Id: typePtrPsbV2Uint, Name: "ptr_pc_psb_v2_uint_t"},
		BlockContextIdPtrPsbV3Uint:              {Id: typePtrPsbV3Uint, Name: "ptr_pc_psb_v3_uint_t"},
		BlockContextIdPtrPsbV4Uint:              {Id: typePtrPsbV4Uint, Name: "ptr_pc_psb_v4_uint_t"},
		BlockContextIdPtrFnUint:                 {Id: typePtrFnUint, Name: "ptr_fn_uint_t"},
		BlockContextIdPosOut:                    {Id: typePosOut, Name: "pos_out_t"},
		BlockContextIdFragDepthOut:              {Id: typeFragDepthOut, Name: "frag_depth_out_t"},
		BlockContextIdZeroVec4:                  {Id: typeZeroVec4, Name: "zero_vec4_t"},
		BlockContextIdBindlessTextures:          {Id: typeBindlessTexturesVar, Name: "bindless_textures_var_t"},
		BlockContextIdPcVar:                     {Id: typePcVar, Name: "pc_var_t"},
		BlockContextIdGlsl:                      {Id: typeGLSL, Name: "glsl_t"},
		BlockContextIdSubgroupLocalInvocationId: {Id: typeSubgroupLocalInvocationId, Name: "subgroup_local_invocation_id_t"},
		BlockContextIdVertexIndex:               {Id: typeVertexIndex, Name: "vertex_index_t"},
		BlockContextIdInstanceIndex:             {Id: typeInstanceIndex, Name: "instance_index_t"},
		BlockContextIdTypeImageBuffer:           {Id: typeTexelBuffer, Name: "image_buffer_t"},
	}
	for i, id := range idTexelBufferVars {
		ids[BlockContextIdTexelBuffer0+BlockContextId(i)] = SpirvBlockContextUsedId{Id: id, Name: fmt.Sprintf("texel_buffer_%d", i)}
	}
	for i, id := range idColorOuts {
		ids[BlockContextIdColorOut0+BlockContextId(i)] = SpirvBlockContextUsedId{Id: id, Name: fmt.Sprintf("color_out_%d", i)}
	}
	for i, id := range idParamOuts {
		ids[BlockContextIdParamOut0+BlockContextId(i)] = SpirvBlockContextUsedId{Id: id, Name: fmt.Sprintf("param_out_%d", i)}
	}

	// Pre-allocate SPIR-V labels ID for GCN CFG blocks.
	labelIds := make([]uint32, len(shader.Cfg.Blocks))
	for i := range shader.Cfg.Blocks {
		labelIds[i] = b.AllocId()
		b.EmitName(labelIds[i], fmt.Sprintf("bb_%d", i))
	}

	// Prepare block context with all GCN and our internal IDs.
	blockContext := SpirvBlockContext{
		Stage:          shader.Stage,
		LabelIds:       labelIds,
		Ids:            ids,
		ConstIds:       constIds,
		GcnSgprArrayId: idSgprArrayVar,
		GcnVgprArrayId: idVgprArrayVar,
		GcnSpecialIds:  gcnSpecialIds,
		GcnConstIds:    gcnConstIds,
	}

	// Function body.
	b.EmitFunction(typeVoid, SpvFunctionControlNone, idFnType, idMain)

	// Emit reachable blocks in reverse post-order (entry block first).
	rpoBlockIds := shader.Cfg.ReversePostOrder()
	emittedBlockIds := make([]bool, len(shader.Cfg.Blocks))

	// Emit reachable blocks.
	for _, blockId := range rpoBlockIds {
		emitBlock(b, &shader.Cfg.Blocks[blockId], &blockContext)
		emittedBlockIds[blockId] = true
	}

	// Emit any unreachable blocks.
	for i := range shader.Cfg.Blocks {
		if !emittedBlockIds[i] {
			b.EmitLabel(labelIds[i])
			b.EmitUnreachable()
		}
	}

	// Emit names for internal IDs.
	for _, c := range blockContext.Ids {
		if c.Id == 0 {
			continue
		}
		b.EmitName(c.Id, c.Name)
	}

	// Emit deferred constants.
	for _, c := range blockContext.GcnConstIds {
		if !c.Used {
			continue
		}
		b.EmitDeferredConstantUint(typeUint, c.Id, c.Value)
		b.EmitName(c.Id, fmt.Sprintf("gcn_const_%s", c.Name))
	}
	for i, c := range blockContext.ConstIds {
		if !c.Used {
			continue
		}
		switch {
		case i >= ConstIdxFloat0:
			b.EmitDeferredConstantFloat(typeFloat, c.Id, math.Float32frombits(c.Value))
		case i >= ConstIdx64Uint0:
			b.EmitDeferredConstantUint64(typeUint64, c.Id, c.Value64)
		default:
			b.EmitDeferredConstantUint(typeUint, c.Id, c.Value)
		}
		b.EmitName(c.Id, fmt.Sprintf("const_%s", c.Name))
	}

	// Emit deferred local variables for used registers.
	for i, c := range blockContext.GcnSpecialIds {
		if i == GcnSpecIdxReserved || !c.Used {
			continue // reserved.
		}
		b.EmitDeferredLocalVariable(typePtrFnUint, c.Id)
		b.EmitName(c.Id, c.Name)
	}

	// Andddd we're done :)
	b.EmitFunctionEnd()

	return &SpirvShader{
		Address: shader.Address,
		Stage:   shader.Stage,
		Code:    b.Assemble(),
	}, nil
}
