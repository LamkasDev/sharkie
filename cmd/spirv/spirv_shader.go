package spirv

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
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
	b.EmitCapability(spec.SpvCapShader)
	b.EmitCapability(spec.SpvCapAddresses)
	b.EmitCapability(spec.SpvCapInt64)
	b.EmitCapability(spec.SpvCapSampled1D)
	b.EmitCapability(spec.SpvCapSampledBuffer)
	b.EmitCapability(spec.SpvCapImageQuery)
	b.EmitCapability(spec.SpvCapGroupNonUniformBallot)
	b.EmitCapability(spec.SpvCapSubgroupBallotKHR)
	b.EmitCapability(spec.SpvCapRuntimeDescriptorArray)
	b.EmitCapability(spec.SpvCapPhysicalStorageBufferAddresses)
	b.EmitExtension("SPV_KHR_physical_storage_buffer")
	b.EmitExtension("SPV_KHR_shader_ballot")
	b.EmitExtension("SPV_EXT_descriptor_indexing")
	b.EmitExtension("SPV_KHR_non_semantic_info")
	typeGLSL := b.EmitExtInstImport("GLSL.std.450")
	typeDebugPrintf := b.EmitExtInstImport("NonSemantic.DebugPrintf")
	b.EmitMemoryModel(spec.SpvAddrModelPhysicalStorageBuffer64, spec.SpvMemModelGLSL450)

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
	typePtrUniformSampledImage2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage2d)
	typePtrUniformBindlessArray2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeBindlessArray2d)

	// Built-ins.
	idTrue := b.EmitConstantTrue(typeBool)
	idFalse := b.EmitConstantFalse(typeBool)

	typePtrInputUint := b.EmitTypePointer(spec.SpvStorageInput, typeUint)
	typeSubgroupLocalInvocationId := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitDecorate(typeSubgroupLocalInvocationId, spec.SpvDecorationBuiltIn, spec.SpvBuiltInSubgroupLocalInvocationId)

	typeVertexIndex := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitName(typeVertexIndex, "vertex_index")
	b.EmitDecorate(typeVertexIndex, spec.SpvDecorationBuiltIn, spec.SpvBuiltInVertexIndex)

	typeInstanceIndex := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitName(typeInstanceIndex, "instance_index")
	b.EmitDecorate(typeInstanceIndex, spec.SpvDecorationBuiltIn, spec.SpvBuiltInInstanceIndex)

	typePtrInputV4F := b.EmitTypePointer(spec.SpvStorageInput, typeV4Float)
	typeFragCoord := b.EmitVariable(typePtrInputV4F, spec.SpvStorageInput)
	b.EmitName(typeFragCoord, "frag_coord")
	b.EmitDecorate(typeFragCoord, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFragCoord)

	// Push constant types.
	typePtrPsbUint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeUint)
	typePtrPsbV2Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV2Uint)
	typePtrPsbV3Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV3Uint)
	typePtrPsbV4Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV4Uint)
	b.EmitDecorate(typePtrPsbUint, spec.SpvDecorationArrayStride, 4)
	b.EmitDecorate(typePtrPsbV2Uint, spec.SpvDecorationArrayStride, 8)
	b.EmitDecorate(typePtrPsbV3Uint, spec.SpvDecorationArrayStride, 12)
	b.EmitDecorate(typePtrPsbV4Uint, spec.SpvDecorationArrayStride, 16)

	// Push constant struct.
	// struct StubPushConstants {
	// 	PhysicalStorageBuffer uint* UserDataAddress;
	//  uint64_t OnionMemoryBaseAddress;
	//  uint64_t GarlicMemoryBaseAddress;
	//  uint_t TexelBuffer0FormatSize;
	//  uint_t TexelBuffer1FormatSize;
	//  uint_t TexelBuffer2FormatSize;
	//  uint_t TexelBuffer3FormatSize;
	//  uint_t TexelBuffer0FormatStride;
	//  uint_t TexelBuffer1FormatStride;
	//  uint_t TexelBuffer2FormatStride;
	//  uint_t TexelBuffer3FormatStride;
	// }
	typePc := b.EmitTypeStruct(typePtrPsbUint, typeUint64, typeUint64,
		typeUint, typeUint, typeUint, typeUint,
		typeUint, typeUint, typeUint, typeUint,
	)
	typePtrPcPsbUint := b.EmitTypePointer(spec.SpvStoragePushConstant, typePtrPsbUint)
	typePtrPcUint64 := b.EmitTypePointer(spec.SpvStoragePushConstant, typeUint64)
	typePtrPcUint := b.EmitTypePointer(spec.SpvStoragePushConstant, typeUint)

	// Annotations for the push-constants.
	b.EmitDecorate(typePc, spec.SpvDecorationBlock)
	b.EmitMemberDecorate(typePc, 0, spec.SpvDecorationOffset, 0)
	b.EmitMemberDecorate(typePc, 1, spec.SpvDecorationOffset, 8)
	b.EmitMemberDecorate(typePc, 2, spec.SpvDecorationOffset, 16)

	b.EmitMemberDecorate(typePc, 3, spec.SpvDecorationOffset, 24)
	b.EmitMemberDecorate(typePc, 4, spec.SpvDecorationOffset, 28)
	b.EmitMemberDecorate(typePc, 5, spec.SpvDecorationOffset, 32)
	b.EmitMemberDecorate(typePc, 6, spec.SpvDecorationOffset, 36)

	b.EmitMemberDecorate(typePc, 7, spec.SpvDecorationOffset, 40)
	b.EmitMemberDecorate(typePc, 8, spec.SpvDecorationOffset, 44)
	b.EmitMemberDecorate(typePc, 9, spec.SpvDecorationOffset, 48)
	b.EmitMemberDecorate(typePc, 10, spec.SpvDecorationOffset, 52)

	// Global push-constant variable.
	typePtrPc := b.EmitTypePointer(spec.SpvStoragePushConstant, typePc)
	pcVar := b.EmitVariable(typePtrPc, spec.SpvStoragePushConstant)
	b.EmitName(pcVar, "push_constants")
	b.EmitDecorate(pcVar, spec.SpvDecorationAliasedPointer)

	// Bindless textures.
	typeBindlessTexturesVar := b.EmitVariable(typePtrUniformBindlessArray2d, spec.SpvStorageUniformConstant)
	b.EmitName(typeBindlessTexturesVar, "bindless_textures")
	b.EmitDecorate(typeBindlessTexturesVar, spec.SpvDecorationDescriptorSet, 0)
	b.EmitDecorate(typeBindlessTexturesVar, spec.SpvDecorationBinding, 0)

	// Texel buffers (Set 1, Binding 0..3).
	var typeTexelBuffer SpirvId
	var idTexelBufferVars [4]SpirvId
	if shader.Stage == GcnShaderStageVertex {
		typeTexelBuffer = b.EmitTypeImage(typeFloat, 5, 0, 0, 0, 1, 0) // Dim=5 (Buffer)
		typePtrUniformTexelBuffer := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeTexelBuffer)
		for i := range 4 {
			idTexelBufferVars[i] = b.EmitVariable(typePtrUniformTexelBuffer, spec.SpvStorageUniformConstant)
			b.EmitName(idTexelBufferVars[i], fmt.Sprintf("texel_buffer_%d", i))
			b.EmitDecorate(idTexelBufferVars[i], spec.SpvDecorationDescriptorSet, 1)
			b.EmitDecorate(idTexelBufferVars[i], spec.SpvDecorationBinding, uint32(i))
		}
	}

	// Stage-specific outputs.
	idPtrOutF := b.EmitTypePointer(spec.SpvStorageOutput, typeFloat)
	idPtrOutV4 := b.EmitTypePointer(spec.SpvStorageOutput, typeV4Float)

	interfaceIds := []SpirvId{typeSubgroupLocalInvocationId, typeVertexIndex, typeInstanceIndex}
	var typePosOut, typeFragDepthOut SpirvId
	var idColorOuts [8]SpirvId
	var idParamOuts [32]SpirvId
	var typeZeroVec4 SpirvId
	switch shader.Stage {
	case GcnShaderStageVertex:
		typePosOut = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
		b.EmitName(typePosOut, "pos_out")
		b.EmitDecorate(typePosOut, spec.SpvDecorationBuiltIn, spec.SpvBuiltInPosition)

		interfaceIds = append(interfaceIds, typePosOut)
		for i := range idParamOuts {
			idParamOuts[i] = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
			b.EmitDecorate(idParamOuts[i], spec.SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idParamOuts[i])
		}
	case GcnShaderStageFragment:
		for i := range idColorOuts {
			idColorOuts[i] = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
			b.EmitDecorate(idColorOuts[i], spec.SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idColorOuts[i])
		}

		typeFragDepthOut = b.EmitVariable(idPtrOutF, spec.SpvStorageOutput)
		b.EmitName(typeFragDepthOut, "frag_depth_out")
		b.EmitDecorate(typeFragDepthOut, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFragDepth)
		interfaceIds = append(interfaceIds, typeFragDepthOut, typeFragCoord)

		// Constant zero vec4 written on exit.
		idZeroF := b.EmitConstantFloat(typeFloat, 0.0)
		idOneF := b.EmitConstantFloat(typeFloat, 1.0)
		typeZeroVec4 = b.EmitConstantComposite(typeV4Float, idZeroF, idZeroF, idZeroF, idOneF)
		b.EmitName(typeZeroVec4, "zero_vec4")
	}

	// Entry point.
	idMain := b.AllocId()
	b.EmitEntryPoint(spec.GncStageToSpvExecModel[shader.Stage], idMain, "main", interfaceIds...)

	// Execution modes.
	switch shader.Stage {
	case GcnShaderStageFragment:
		b.EmitExecutionMode(idMain, spec.SpvExecModeOriginUpperLeft)
	case GcnShaderStageCompute:
		numThreads := ctx.NumThreads
		for i := range numThreads {
			if numThreads[i] == 0 {
				numThreads[i] = 1
			}
		}
		b.EmitExecutionMode(idMain, spec.SpvExecModeLocalSize, numThreads[0], numThreads[1], numThreads[2])
	}

	// Register GCN SGPRs and VGPRs.
	typePtrFnUint := b.EmitTypePointer(spec.SpvStorageFunction, typeUint)
	idSgprCount := b.EmitConstantUint(typeUint, 104)
	idVgprCount := b.EmitConstantUint(typeUint, 256)
	typeSgprArray := b.EmitTypeArray(typeUint, idSgprCount)
	typeVgprArray := b.EmitTypeArray(typeUint, idVgprCount)
	typePtrSgprArray := b.EmitTypePointer(spec.SpvStorageFunction, typeSgprArray)
	typePtrVgprArray := b.EmitTypePointer(spec.SpvStorageFunction, typeVgprArray)

	idSgprArrayVar := b.AllocId()
	idVgprArrayVar := b.AllocId()
	b.EmitName(idSgprArrayVar, "sgprs")
	b.EmitName(idVgprArrayVar, "vgprs")
	b.EmitDeferredLocalVariable(typePtrSgprArray, idSgprArrayVar)
	b.EmitDeferredLocalVariable(typePtrVgprArray, idVgprArrayVar)

	// GCN special registers.
	var gcnSpecialIds [27]SpirvUsedId
	gcnSpecialIds[GcnSpecIdFlatScrLo] = SpirvUsedId{Id: b.AllocId(), Name: "flat_scr_lo"}
	gcnSpecialIds[GcnSpecIdFlatScrHi] = SpirvUsedId{Id: b.AllocId(), Name: "flat_scr_hi"}
	gcnSpecialIds[GcnSpecIdVccLo] = SpirvUsedId{Id: b.AllocId(), Name: "vcc_lo"}
	gcnSpecialIds[GcnSpecIdVccHi] = SpirvUsedId{Id: b.AllocId(), Name: "vcc_hi"}
	gcnSpecialIds[GcnSpecIdTbaLo] = SpirvUsedId{Id: b.AllocId(), Name: "tba_lo"}
	gcnSpecialIds[GcnSpecIdTbaHi] = SpirvUsedId{Id: b.AllocId(), Name: "tba_hi"}
	gcnSpecialIds[GcnSpecIdTmaLo] = SpirvUsedId{Id: b.AllocId(), Name: "tma_lo"}
	gcnSpecialIds[GcnSpecIdTmaHi] = SpirvUsedId{Id: b.AllocId(), Name: "tma_hi"}
	for i := range SpirvId(12) {
		gcnSpecialIds[GcnSpecIdTtmp0+i] = SpirvUsedId{Id: b.AllocId(), Name: fmt.Sprintf("ttmp%d", i)}
	}
	gcnSpecialIds[GcnSpecIdM0] = SpirvUsedId{Id: b.AllocId(), Name: "m0"}
	gcnSpecialIds[GcnSpecIdExecLo] = SpirvUsedId{Id: b.AllocId(), Name: "exec_lo"}
	gcnSpecialIds[GcnSpecIdExecHi] = SpirvUsedId{Id: b.AllocId(), Name: "exec_hi"}
	gcnSpecialIds[GcnSpecIdVccz] = SpirvUsedId{Id: b.AllocId(), Name: "vccz"}
	gcnSpecialIds[GcnSpecIdExecz] = SpirvUsedId{Id: b.AllocId(), Name: "execz"}
	gcnSpecialIds[GcnSpecIdScc] = SpirvUsedId{Id: b.AllocId(), Name: "scc"}

	// GCN inline constants.
	var gcnConstIds [120]SpirvUsedId
	for i, id := range gcnConstIds {
		usedId := SpirvUsedId{Id: b.AllocId()}
		switch {
		case id.Id == GcnConstId0:
			usedId.Value, usedId.Name = 0, "0"
		case id.Id >= GcnConstIdInt1 && id.Id <= GcnConstIdInt64:
			usedId.Value, usedId.Name = uint32(i), fmt.Sprint(i)
		case id.Id >= GcnConstIdIntNeg1 && id.Id <= GcnConstIdIntNeg16:
			v := int32(-(i - int(GcnConstIdInt64)))
			usedId.Value, usedId.Name = uint32(v), fmt.Sprint(v)
		case id.Id == GcnConstIdFloat05:
			usedId.Value, usedId.Name = math.Float32bits(0.5), "0.5"
		case id.Id == GcnConstIdFloatNeg05:
			usedId.Value, usedId.Name = math.Float32bits(-0.5), "-0.5"
		case id.Id == GcnConstIdFloat10:
			usedId.Value, usedId.Name = math.Float32bits(1.0), "1.0"
		case id.Id == GcnConstIdFloatNeg10:
			usedId.Value, usedId.Name = math.Float32bits(-1.0), "-1.0"
		case id.Id == GcnConstIdFloat20:
			usedId.Value, usedId.Name = math.Float32bits(2.0), "2.0"
		case id.Id == GcnConstIdFloatNeg20:
			usedId.Value, usedId.Name = math.Float32bits(-2.0), "-2.0"
		case id.Id == GcnConstIdFloat40:
			usedId.Value, usedId.Name = math.Float32bits(4.0), "4.0"
		case id.Id == GcnConstIdFloatNeg40:
			usedId.Value, usedId.Name = math.Float32bits(-4.0), "-4.0"
		}
		gcnConstIds[i] = usedId
	}

	// Internal inline constants.
	constIds := map[SpirvId]SpirvUsedId{}
	for i := range SpirvId(256) {
		constIds[i] = SpirvUsedId{Id: b.AllocId(), Value: uint32(i), Name: fmt.Sprint(i)}
	}
	constIds[ConstIdUint3FFF] = SpirvUsedId{Id: b.AllocId(), Value: 0x3FFF, Name: "0x3FFF"}
	constIds[ConstIdUintFFFF] = SpirvUsedId{Id: b.AllocId(), Value: 0xFFFF, Name: "0xFFFF"}
	constIds[ConstIdUint11111111] = SpirvUsedId{Id: b.AllocId(), Value: 0x11111111, Name: "0x11111111"}
	constIds[ConstIdUintFFFFFFFF] = SpirvUsedId{Id: b.AllocId(), Value: 0xFFFFFFFF, Name: "0xFFFFFFFF"}
	constIds[ConstId64Uint0] = SpirvUsedId{Id: b.AllocId(), Value64: 0, Name: "64_0"}
	constIds[ConstId64Uint1] = SpirvUsedId{Id: b.AllocId(), Value64: 1, Name: "64_1"}
	constIds[ConstId64Uint32] = SpirvUsedId{Id: b.AllocId(), Value64: 32, Name: "64_32"}
	constIds[ConstIdFloat1] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(1.0), Name: "1.0"}
	constIds[ConstIdFloat0] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(0.0), Name: "0.0"}

	// Prepare internal IDs.
	ids := map[SpirvId]SpirvUsedId{
		BlockContextIdFalse:       {Id: idFalse, Name: "false"},
		BlockContextIdTrue:        {Id: idTrue, Name: "true"},
		BlockContextIdTypeBool:    {Id: typeBool, Name: "bool_t"},
		BlockContextIdTypeFloat:   {Id: typeFloat, Name: "float_t"},
		BlockContextIdTypeInt:     {Id: typeInt, Name: "int_t"},
		BlockContextIdTypeUint:    {Id: typeUint, Name: "uint_t"},
		BlockContextIdTypeUint64:  {Id: typeUint64, Name: "uint64_t"},
		BlockContextIdTypeInt64:   {Id: typeInt64, Name: "int64_t"},
		BlockContextIdTypeVoid:    {Id: typeVoid, Name: "void_t"},
		BlockContextIdDebugPrintf: {Id: typeDebugPrintf, Name: "debug_printf_t"},
		BlockContextIdTypeV2Float: {Id: typeV2Float, Name: "v2float_t"},

		BlockContextIdTypeV4Float:               {Id: typeV4Float, Name: "v4float_t"},
		BlockContextIdTypeV2Uint:                {Id: typeV2Uint, Name: "v2uint_t"},
		BlockContextIdTypeV3Uint:                {Id: typeV3Uint, Name: "v3uint_t"},
		BlockContextIdTypeV4Uint:                {Id: typeV4Uint, Name: "v4uint_t"},
		BlockContextIdTypeSampledImage:          {Id: typeSampledImage2d, Name: "sampled_image_2d_t"},
		BlockContextIdPtrUniformSampledImage:    {Id: typePtrUniformSampledImage2d, Name: "ptr_uniform_sampled_image_2d_t"},
		BlockContextIdPtrPcPsbUint:              {Id: typePtrPcPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPcUint:                 {Id: typePtrPcUint, Name: "ptr_pc_uint_t"},
		BlockContextIdPtrPcUint64:               {Id: typePtrPcUint64, Name: "ptr_pc_uint64_t"},
		BlockContextIdPtrPsbUint:                {Id: typePtrPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPsbV2Uint:              {Id: typePtrPsbV2Uint, Name: "ptr_pc_psb_v2_uint_t"},
		BlockContextIdPtrPsbV3Uint:              {Id: typePtrPsbV3Uint, Name: "ptr_pc_psb_v3_uint_t"},
		BlockContextIdPtrPsbV4Uint:              {Id: typePtrPsbV4Uint, Name: "ptr_pc_psb_v4_uint_t"},
		BlockContextIdPtrFnUint:                 {Id: typePtrFnUint, Name: "ptr_fn_uint_t"},
		BlockContextIdPosOut:                    {Id: typePosOut, Name: "pos_out_t"},
		BlockContextIdFragDepthOut:              {Id: typeFragDepthOut, Name: "frag_depth_out_t"},
		BlockContextIdZeroVec4:                  {Id: typeZeroVec4, Name: "zero_vec4_t"},
		BlockContextIdBindlessTextures:          {Id: typeBindlessTexturesVar, Name: "bindless_textures_var_t"},
		BlockContextIdPcVar:                     {Id: pcVar, Name: "pc_var_t"},
		BlockContextIdGlsl:                      {Id: typeGLSL, Name: "glsl_t"},
		BlockContextIdSubgroupLocalInvocationId: {Id: typeSubgroupLocalInvocationId, Name: "subgroup_local_invocation_id_t"},
		BlockContextIdVertexIndex:               {Id: typeVertexIndex, Name: "vertex_index_t"},
		BlockContextIdInstanceIndex:             {Id: typeInstanceIndex, Name: "instance_index_t"},
		BlockContextIdFragCoord:                 {Id: typeFragCoord, Name: "frag_coord_t"},
		BlockContextIdTypeImageBuffer:           {Id: typeTexelBuffer, Name: "image_buffer_t"},
	}
	for i, id := range idTexelBufferVars {
		ids[BlockContextIdTexelBuffer0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("texel_buffer_%d", i)}
	}
	for i, id := range idColorOuts {
		ids[BlockContextIdColorOut0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("color_out_%d", i)}
	}
	for i, id := range idParamOuts {
		ids[BlockContextIdParamOut0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("param_out_%d", i)}
	}

	// Pre-allocate SPIR-V labels ID for GCN CFG blocks.
	labelIds := make([]SpirvId, len(shader.Cfg.Blocks))
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
	b.EmitFunction(typeVoid, idMain, spec.SpvFunctionControlNone, idFnType)

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
		case i >= ConstIdFloat0:
			b.EmitDeferredConstantFloat(typeFloat, c.Id, math.Float32frombits(c.Value))
		case i >= ConstId64Uint0:
			b.EmitDeferredConstantUint64(typeUint64, c.Id, c.Value64)
		default:
			b.EmitDeferredConstantUint(typeUint, c.Id, c.Value)
		}
		b.EmitName(c.Id, fmt.Sprintf("const_%s", c.Name))
	}

	// Emit deferred local variables for used registers.
	for _, c := range blockContext.GcnSpecialIds {
		if c.Id == GcnSpecIdReserved || !c.Used {
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
