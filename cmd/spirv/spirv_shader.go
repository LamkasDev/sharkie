package spirv

import (
	"fmt"
	"math"

	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

type SpirvShaderContext struct {
	ThreadX uint32
	ThreadY uint32
	ThreadZ uint32

	PsInputAddress  uint32
	PsInputControls [32]uint32
}

type SpirvShader struct {
	Stage     GcnShaderStage
	Address   uintptr
	Code      []uint32
	Resources []SpirvShaderResource
}

func NewSpirvShader(shader *GcnShader, ctx SpirvShaderContext) (*SpirvShader, error) {
	resources := AnalyzeResources(shader)
	b := NewSpvBuilder()

	// Capabilities.
	b.EmitCapability(spec.SpvCapShader)
	b.EmitCapability(spec.SpvCapInt64)
	b.EmitCapability(spec.SpvCapSampled1D)
	b.EmitCapability(spec.SpvCapSampledBuffer)
	b.EmitCapability(spec.SpvCapImageQuery)
	if shader.Stage == GcnShaderStageFragment {
		b.EmitCapability(spec.SpvCapInterpolationFunction)
	}
	b.EmitCapability(spec.SpvCapGroupNonUniformBallot)
	b.EmitCapability(spec.SpvCapSubgroupBallotKHR)
	b.EmitCapability(spec.SpvCapRuntimeDescriptorArray)
	b.EmitCapability(spec.SpvCapPhysicalStorageBufferAddresses)
	b.EmitCapability(spec.SpvCapShaderNonUniform)
	b.EmitCapability(spec.SpvCapSampledImageArrayNonUniformIndexing)
	b.EmitCapability(spec.SpvCapStorageImageReadWithoutFormat)
	b.EmitCapability(spec.SpvCapStorageImageWriteWithoutFormat)
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

	typeV2Int := b.EmitTypeVector(typeInt, 2)
	typeV4Int := b.EmitTypeVector(typeInt, 4)
	typeV2Uint := b.EmitTypeVector(typeUint, 2)
	typeV3Uint := b.EmitTypeVector(typeUint, 3)
	typeV4Uint := b.EmitTypeVector(typeUint, 4)
	typeStructUintUint := b.EmitTypeStruct(typeUint, typeUint)

	typeFloat := b.EmitTypeFloat(32)
	typeV2Float := b.EmitTypeVector(typeFloat, 2)
	typeV4Float := b.EmitTypeVector(typeFloat, 4)

	typeImage2d := b.EmitTypeImage(typeFloat, 1, 0, 0, 0, 1, 0)
	typeSampledImage2d := b.EmitTypeSampledImage(typeImage2d)
	typeBindlessArray2d := b.EmitTypeRuntimeArray(typeSampledImage2d)
	typePtrUniformSampledImage2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage2d)
	typePtrUniformBindlessArray2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeBindlessArray2d)

	typeStorageImage2d := b.EmitTypeImage(typeFloat, 1, 0, 0, 0, 2, 0)
	typeBindlessStorageArray2d := b.EmitTypeRuntimeArray(typeStorageImage2d)
	typePtrUniformStorageImage2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStorageImage2d)
	typePtrUniformBindlessStorageArray2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeBindlessStorageArray2d)

	// Built-ins.
	idTrue := b.EmitConstantTrue(typeBool)
	idFalse := b.EmitConstantFalse(typeBool)

	typePtrInputUint := b.EmitTypePointer(spec.SpvStorageInput, typeUint)
	typeSubgroupLocalInvocationId := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitDecorate(typeSubgroupLocalInvocationId, spec.SpvDecorationBuiltIn, spec.SpvBuiltInSubgroupLocalInvocationId)
	if shader.Stage == GcnShaderStageFragment {
		b.EmitDecorate(typeSubgroupLocalInvocationId, spec.SpvDecorationFlat)
	}

	typeVertexIndex := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitName(typeVertexIndex, "vertex_index")
	b.EmitDecorate(typeVertexIndex, spec.SpvDecorationBuiltIn, spec.SpvBuiltInVertexIndex)
	if shader.Stage == GcnShaderStageFragment {
		b.EmitDecorate(typeVertexIndex, spec.SpvDecorationFlat)
	}

	typeInstanceIndex := b.EmitVariable(typePtrInputUint, spec.SpvStorageInput)
	b.EmitName(typeInstanceIndex, "instance_index")
	b.EmitDecorate(typeInstanceIndex, spec.SpvDecorationBuiltIn, spec.SpvBuiltInInstanceIndex)
	if shader.Stage == GcnShaderStageFragment {
		b.EmitDecorate(typeInstanceIndex, spec.SpvDecorationFlat)
	}

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
	//  uint_t UserSgprCount;
	//  uint_t ShaderRsrc2;
	//  uint_t VteControl;
	//  uint_t ClipControl;
	//  float_t GbHorzClipAdj;
	//  float_t GbVertClipAdj;
	// }
	typePc := b.EmitTypeStruct(
		typePtrPsbUint, typeUint64, typeUint64,
		typeUint, typeUint, typeUint, typeUint,
		typeFloat, typeFloat,
	)
	typePtrPcPsbUint := b.EmitTypePointer(spec.SpvStoragePushConstant, typePtrPsbUint)
	typePtrPcUint64 := b.EmitTypePointer(spec.SpvStoragePushConstant, typeUint64)
	typePtrPcUint := b.EmitTypePointer(spec.SpvStoragePushConstant, typeUint)
	typePtrPcFloat := b.EmitTypePointer(spec.SpvStoragePushConstant, typeFloat)

	// Annotations for the push-constants.
	baseOffset := uint32(0)
	if shader.Stage == GcnShaderStageFragment {
		baseOffset = PushConstantsSize
	}
	b.EmitDecorate(typePc, spec.SpvDecorationBlock)
	b.EmitMemberDecorate(typePc, 0, spec.SpvDecorationOffset, baseOffset+0)
	b.EmitMemberDecorate(typePc, 1, spec.SpvDecorationOffset, baseOffset+8)
	b.EmitMemberDecorate(typePc, 2, spec.SpvDecorationOffset, baseOffset+16)

	b.EmitMemberDecorate(typePc, 3, spec.SpvDecorationOffset, baseOffset+24)
	b.EmitMemberDecorate(typePc, 4, spec.SpvDecorationOffset, baseOffset+28)
	b.EmitMemberDecorate(typePc, 5, spec.SpvDecorationOffset, baseOffset+32)
	b.EmitMemberDecorate(typePc, 6, spec.SpvDecorationOffset, baseOffset+36)

	b.EmitMemberDecorate(typePc, 7, spec.SpvDecorationOffset, baseOffset+40)
	b.EmitMemberDecorate(typePc, 8, spec.SpvDecorationOffset, baseOffset+44)

	// Global push-constant variable.
	typePtrPc := b.EmitTypePointer(spec.SpvStoragePushConstant, typePc)
	pcVar := b.EmitVariable(typePtrPc, spec.SpvStoragePushConstant)
	b.EmitName(pcVar, "push_constants")
	b.EmitDecorate(pcVar, spec.SpvDecorationAliasedPointer)

	// Bindless textures.
	typeBindlessTexturesVar := b.EmitVariable(typePtrUniformBindlessArray2d, spec.SpvStorageUniformConstant)
	b.EmitName(typeBindlessTexturesVar, "bindless_textures")
	b.EmitDecorate(typeBindlessTexturesVar, spec.SpvDecorationDescriptorSet, DescriptorSetSlotBindless)
	b.EmitDecorate(typeBindlessTexturesVar, spec.SpvDecorationBinding, 0)

	// Bindless storage textures.
	typeBindlessStorageTexturesVar := b.EmitVariable(typePtrUniformBindlessStorageArray2d, spec.SpvStorageUniformConstant)
	b.EmitName(typeBindlessStorageTexturesVar, "bindless_storage_textures")
	b.EmitDecorate(typeBindlessStorageTexturesVar, spec.SpvDecorationDescriptorSet, DescriptorSetSlotBindless)
	b.EmitDecorate(typeBindlessStorageTexturesVar, spec.SpvDecorationBinding, 1)

	// Global descriptor map.
	typePtrStorageUint := b.EmitTypePointer(spec.SpvStorageStorageBuffer, typeUint)
	typeDescriptorMapArray := b.EmitTypeRuntimeArray(typeUint)
	b.EmitDecorate(typeDescriptorMapArray, spec.SpvDecorationArrayStride, 4)
	typeDescriptorMap := b.EmitTypeStruct(typeDescriptorMapArray)
	b.EmitDecorate(typeDescriptorMap, spec.SpvDecorationBlock)
	b.EmitMemberDecorate(typeDescriptorMap, 0, spec.SpvDecorationOffset, 0)
	typePtrStorageDescriptorMap := b.EmitTypePointer(spec.SpvStorageStorageBuffer, typeDescriptorMap)
	idDescriptorMapVar := b.EmitVariable(typePtrStorageDescriptorMap, spec.SpvStorageStorageBuffer)
	b.EmitName(idDescriptorMapVar, "global_descriptor_map")
	b.EmitDecorate(idDescriptorMapVar, spec.SpvDecorationDescriptorSet, DescriptorSetSlotDiscovery)
	b.EmitDecorate(idDescriptorMapVar, spec.SpvDecorationBinding, 0)
	b.EmitDecorate(idDescriptorMapVar, spec.SpvDecorationCoherent)
	b.EmitDecorate(idDescriptorMapVar, spec.SpvDecorationVolatile)

	// Missing resource buffer.
	typeMissingDescriptor := b.EmitTypeStruct(
		typeUint, typeUint, typeUint, typeUint,
		typeUint, typeUint, typeUint, typeUint,
		typeUint, typeUint, typeUint, typeUint,
	)
	for i := range uint32(12) {
		b.EmitMemberDecorate(typeMissingDescriptor, i, spec.SpvDecorationOffset, i*4)
	}
	typeMissingDescriptorsArray := b.EmitTypeRuntimeArray(typeMissingDescriptor)
	b.EmitDecorate(typeMissingDescriptorsArray, spec.SpvDecorationArrayStride, 48)
	typeMissingResourceBuffer := b.EmitTypeStruct(typeUint, typeMissingDescriptorsArray)
	b.EmitDecorate(typeMissingResourceBuffer, spec.SpvDecorationBlock)
	b.EmitMemberDecorate(typeMissingResourceBuffer, 0, spec.SpvDecorationOffset, 0)
	b.EmitMemberDecorate(typeMissingResourceBuffer, 1, spec.SpvDecorationOffset, 16)
	typePtrStorageMissingResourceBuffer := b.EmitTypePointer(spec.SpvStorageStorageBuffer, typeMissingResourceBuffer)
	idMissingResourceBufferVar := b.EmitVariable(typePtrStorageMissingResourceBuffer, spec.SpvStorageStorageBuffer)
	b.EmitName(idMissingResourceBufferVar, "missing_resource_buffer")
	b.EmitDecorate(idMissingResourceBufferVar, spec.SpvDecorationDescriptorSet, DescriptorSetSlotDiscovery)
	b.EmitDecorate(idMissingResourceBufferVar, spec.SpvDecorationBinding, 1)
	b.EmitDecorate(idMissingResourceBufferVar, spec.SpvDecorationCoherent)
	b.EmitDecorate(idMissingResourceBufferVar, spec.SpvDecorationVolatile)

	idZeroF := b.EmitConstantFloat(typeFloat, 0.0)
	idOneF := b.EmitConstantFloat(typeFloat, 1.0)
	typeZeroVec4 := b.EmitConstantComposite(typeV4Float, idZeroF, idZeroF, idZeroF, idOneF)
	b.EmitName(typeZeroVec4, "zero_vec4")

	// Stage-specific outputs.
	idPtrOutF := b.EmitTypePointer(spec.SpvStorageOutput, typeFloat)
	idPtrOutV4 := b.EmitTypePointer(spec.SpvStorageOutput, typeV4Float)

	interfaceIds := []SpirvId{typeSubgroupLocalInvocationId, typeVertexIndex, typeInstanceIndex}
	var typePosOut, typeFragDepthOut SpirvId
	var idColorOuts [8]SpirvId
	var idParamOuts [32]SpirvId
	var idParamIns [32]SpirvId
	var idWorkgroupId, idLocalInvocationId SpirvId
	switch shader.Stage {
	case GcnShaderStageVertex:
		typePosOut = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
		b.EmitName(typePosOut, "pos_out")
		b.EmitDecorate(typePosOut, spec.SpvDecorationBuiltIn, spec.SpvBuiltInPosition)

		interfaceIds = append(interfaceIds, typePosOut)
		// TODO: this
		for i := range 16 {
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

		// Use PsInputAddress to determine which parameters to declare.
		var usedParamTypes []uint8
		for i := 0; i < 32; i++ {
			if (ctx.PsInputAddress>>i)&1 != 0 {
				usedParamTypes = append(usedParamTypes, uint8(i))
			}
		}

		for i := range usedParamTypes {
			control := ctx.PsInputControls[i]
			offset := control & 0x3F
			if match := offset&0x20 == 0; match {
				// Vertex shader match found.
				location := offset & 0x1F
				idParamIns[i] = b.EmitVariable(b.EmitTypePointer(spec.SpvStorageInput, typeV4Float), spec.SpvStorageInput)
				b.EmitName(idParamIns[i], fmt.Sprintf("param_in_%d", i))
				b.EmitDecorate(idParamIns[i], spec.SpvDecorationLocation, location)
				if flat := (control>>10)&1 == 1; flat {
					b.EmitDecorate(idParamIns[i], spec.SpvDecorationFlat)
				}
				interfaceIds = append(interfaceIds, idParamIns[i])
			}
		}

		// Handle system inputs based on address bits.
		if (ctx.PsInputAddress>>8)&0xF != 0 {
			typePtrInputV4F := b.EmitTypePointer(spec.SpvStorageInput, typeV4Float)
			typeFragCoord := b.EmitVariable(typePtrInputV4F, spec.SpvStorageInput)
			b.EmitName(typeFragCoord, "frag_coord")
			b.EmitDecorate(typeFragCoord, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFragCoord)
			interfaceIds = append(interfaceIds, typeFragCoord)
		}

		writesDepth := false
		for _, blockId := range shader.Cfg.ReversePostOrder() {
			for _, instr := range shader.Cfg.Blocks[blockId].Instructions {
				if instr.Encoding == gcnSpec.EncEXP {
					if details, ok := instr.Details.(*gcnSpec.ExpDetails); ok && details.Target == 8 {
						writesDepth = true
					}
				}
			}
		}

		if writesDepth {
			typeFragDepthOut = b.EmitVariable(idPtrOutF, spec.SpvStorageOutput)
			b.EmitName(typeFragDepthOut, "frag_depth_out")
			b.EmitDecorate(typeFragDepthOut, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFragDepth)
			interfaceIds = append(interfaceIds, typeFragDepthOut)
		}
	case GcnShaderStageCompute:
		typePtrInputV3Uint := b.EmitTypePointer(spec.SpvStorageInput, typeV3Uint)
		idWorkgroupId = b.EmitVariable(typePtrInputV3Uint, spec.SpvStorageInput)
		b.EmitName(idWorkgroupId, "workgroup_id")
		b.EmitDecorate(idWorkgroupId, spec.SpvDecorationBuiltIn, spec.SpvBuiltInWorkgroupId)
		interfaceIds = append(interfaceIds, idWorkgroupId)

		idLocalInvocationId = b.EmitVariable(typePtrInputV3Uint, spec.SpvStorageInput)
		b.EmitName(idLocalInvocationId, "local_invocation_id")
		b.EmitDecorate(idLocalInvocationId, spec.SpvDecorationBuiltIn, spec.SpvBuiltInLocalInvocationId)
		interfaceIds = append(interfaceIds, idLocalInvocationId)
	}

	// Entry point.
	idMain := b.AllocId()
	b.EmitEntryPoint(spec.GncStageToSpvExecModel[shader.Stage], idMain, "main", interfaceIds...)

	// Execution modes.
	switch shader.Stage {
	case GcnShaderStageFragment:
		b.EmitExecutionMode(idMain, spec.SpvExecModeOriginUpperLeft)
		if typeFragDepthOut != 0 {
			b.EmitExecutionMode(idMain, spec.SpvExecModeDepthReplacing)
		}
	case GcnShaderStageCompute:
		b.EmitExecutionMode(idMain, spec.SpvExecModeLocalSize, ctx.ThreadX, ctx.ThreadY, ctx.ThreadZ)
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
	for i := range gcnConstIds {
		constId := SpirvId(i)
		usedId := SpirvUsedId{Id: b.AllocId()}
		switch {
		case constId == GcnConstId0:
			usedId.Value, usedId.Name = 0, "0"
		case constId >= GcnConstIdInt1 && constId <= GcnConstIdInt64:
			usedId.Value, usedId.Name = uint32(i), fmt.Sprint(i)
		case constId >= GcnConstIdIntNeg1 && constId <= GcnConstIdIntNeg16:
			v := int32(-(i - int(GcnConstIdInt64)))
			usedId.Value, usedId.Name = uint32(v), fmt.Sprint(v)
		case constId == GcnConstIdFloat05:
			usedId.Value, usedId.Name = math.Float32bits(0.5), "0.5"
		case constId == GcnConstIdFloatNeg05:
			usedId.Value, usedId.Name = math.Float32bits(-0.5), "-0.5"
		case constId == GcnConstIdFloat10:
			usedId.Value, usedId.Name = math.Float32bits(1.0), "1.0"
		case constId == GcnConstIdFloatNeg10:
			usedId.Value, usedId.Name = math.Float32bits(-1.0), "-1.0"
		case constId == GcnConstIdFloat20:
			usedId.Value, usedId.Name = math.Float32bits(2.0), "2.0"
		case constId == GcnConstIdFloatNeg20:
			usedId.Value, usedId.Name = math.Float32bits(-2.0), "-2.0"
		case constId == GcnConstIdFloat40:
			usedId.Value, usedId.Name = math.Float32bits(4.0), "4.0"
		case constId == GcnConstIdFloatNeg40:
			usedId.Value, usedId.Name = math.Float32bits(-4.0), "-4.0"
		}
		gcnConstIds[i] = usedId
	}

	// Internal inline constants.
	constIds := map[SpirvId]SpirvUsedId{}
	for i := range SpirvId(257) {
		constIds[i] = SpirvUsedId{Id: b.AllocId(), Value: uint32(i), Name: fmt.Sprint(i)}
	}
	constIds[ConstIdUint3FFF] = SpirvUsedId{Id: b.AllocId(), Value: 0x3FFF, Name: "0x3FFF"}
	constIds[ConstIdUintFFFF] = SpirvUsedId{Id: b.AllocId(), Value: 0xFFFF, Name: "0xFFFF"}
	constIds[ConstIdUint11111111] = SpirvUsedId{Id: b.AllocId(), Value: 0x11111111, Name: "0x11111111"}
	constIds[ConstIdUintFFFFFFFF] = SpirvUsedId{Id: b.AllocId(), Value: 0xFFFFFFFF, Name: "0xFFFFFFFF"}
	constIds[ConstIdUintFFFFFFFFC] = SpirvUsedId{Id: b.AllocId(), Value: 0xFFFFFFFC, Name: "0xFFFFFFFC"}
	constIds[ConstId64Uint0] = SpirvUsedId{Id: b.AllocId(), Value64: 0, Name: "64_0"}
	constIds[ConstId64Uint1] = SpirvUsedId{Id: b.AllocId(), Value64: 1, Name: "64_1"}
	constIds[ConstId64Uint4] = SpirvUsedId{Id: b.AllocId(), Value64: 4, Name: "64_4"}
	constIds[ConstId64Uint8] = SpirvUsedId{Id: b.AllocId(), Value64: 8, Name: "64_8"}
	constIds[ConstId64Uint12] = SpirvUsedId{Id: b.AllocId(), Value64: 12, Name: "64_12"}
	constIds[ConstId64Uint32] = SpirvUsedId{Id: b.AllocId(), Value64: 32, Name: "64_32"}
	constIds[ConstId64UintNot3] = SpirvUsedId{Id: b.AllocId(), Value64: ^uint64(0x3), Name: "64_not_3"}
	constIds[ConstId64UintOnionBaseAddress] = SpirvUsedId{Id: b.AllocId(), Value64: uint64(GlobalAllocator.Base), Name: "onion_base"}
	constIds[ConstId64UintGarlicBaseAddress] = SpirvUsedId{Id: b.AllocId(), Value64: uint64(GlobalGpuAllocator.Base), Name: "garlic_base"}
	constIds[ConstIdFloat0] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(0.0), Name: "0.0"}
	constIds[ConstIdFloat05] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(0.5), Name: "0.5"}
	constIds[ConstIdFloat1] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(1.0), Name: "1.0"}
	constIds[ConstIdFloat2] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(2.0), Name: "2.0"}
	constIds[ConstIdFloat255] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(255.0), Name: "255.0"}
	constIds[ConstIdFloat65535] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(65535.0), Name: "65535.0"}

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
		BlockContextIdTypeV2Int:   {Id: typeV2Int, Name: "v2int_t"},
		BlockContextIdTypeV4Int:   {Id: typeV4Int, Name: "v4int_t"},

		BlockContextIdTypeV4Float:               {Id: typeV4Float, Name: "v4float_t"},
		BlockContextIdTypeV2Uint:                {Id: typeV2Uint, Name: "v2uint_t"},
		BlockContextIdTypeV3Uint:                {Id: typeV3Uint, Name: "v3uint_t"},
		BlockContextIdTypeV4Uint:                {Id: typeV4Uint, Name: "v4uint_t"},
		BlockContextIdTypeStructUintUint:        {Id: typeStructUintUint, Name: "struct_uint_uint_t"},
		BlockContextIdTypeSampledImage:          {Id: typeSampledImage2d, Name: "sampled_image_2d_t"},
		BlockContextIdTypeImage:                 {Id: typeImage2d, Name: "image_2d_t"},
		BlockContextIdPtrUniformSampledImage:    {Id: typePtrUniformSampledImage2d, Name: "ptr_uniform_sampled_image_2d_t"},
		BlockContextIdTypeStorageImage:          {Id: typeStorageImage2d, Name: "storage_image_2d_t"},
		BlockContextIdPtrUniformStorageImage:    {Id: typePtrUniformStorageImage2d, Name: "ptr_uniform_storage_image_2d_t"},
		BlockContextIdPtrPcPsbUint:              {Id: typePtrPcPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPcUint:                 {Id: typePtrPcUint, Name: "ptr_pc_uint_t"},
		BlockContextIdPtrPcUint64:               {Id: typePtrPcUint64, Name: "ptr_pc_uint64_t"},
		BlockContextIdPtrPcFloat:                {Id: typePtrPcFloat, Name: "ptr_pc_float_t"},
		BlockContextIdPtrPsbUint:                {Id: typePtrPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPsbV2Uint:              {Id: typePtrPsbV2Uint, Name: "ptr_pc_psb_v2_uint_t"},
		BlockContextIdPtrPsbV3Uint:              {Id: typePtrPsbV3Uint, Name: "ptr_pc_psb_v3_uint_t"},
		BlockContextIdPtrPsbV4Uint:              {Id: typePtrPsbV4Uint, Name: "ptr_pc_psb_v4_uint_t"},
		BlockContextIdPtrStorageUint:            {Id: typePtrStorageUint, Name: "ptr_storage_uint_t"},
		BlockContextIdPtrFnUint:                 {Id: typePtrFnUint, Name: "ptr_fn_uint_t"},
		BlockContextIdPosOut:                    {Id: typePosOut, Name: "pos_out_t"},
		BlockContextIdFragDepthOut:              {Id: typeFragDepthOut, Name: "frag_depth_out_t"},
		BlockContextIdZeroVec4:                  {Id: typeZeroVec4, Name: "zero_vec4_t"},
		BlockContextIdBindlessTextures:          {Id: typeBindlessTexturesVar, Name: "bindless_textures_var_t"},
		BlockContextIdBindlessStorageTextures:   {Id: typeBindlessStorageTexturesVar, Name: "bindless_storage_textures_var_t"},
		BlockContextIdPcVar:                     {Id: pcVar, Name: "pc_var_t"},
		BlockContextIdGlsl:                      {Id: typeGLSL, Name: "glsl_t"},
		BlockContextIdSubgroupLocalInvocationId: {Id: typeSubgroupLocalInvocationId, Name: "subgroup_local_invocation_id_t"},
		BlockContextIdVertexIndex:               {Id: typeVertexIndex, Name: "vertex_index_t"},
		BlockContextIdInstanceIndex:             {Id: typeInstanceIndex, Name: "instance_index_t"},
		BlockContextIdGlobalDescriptorMap:       {Id: idDescriptorMapVar, Name: "global_descriptor_map_t"},
		BlockContextIdMissingResourceBuffer:     {Id: idMissingResourceBufferVar, Name: "missing_resource_buffer_t"},
		BlockContextIdWorkgroupId:               {Id: idWorkgroupId, Name: "workgroup_id_t"},
		BlockContextIdLocalInvocationId:         {Id: idLocalInvocationId, Name: "local_invocation_id"},
	}
	for i, id := range idColorOuts {
		ids[BlockContextIdColorOut0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("color_out_%d", i)}
	}

	for i, id := range idParamOuts {
		ids[BlockContextIdParamOut0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("param_out_%d", i)}
	}
	for i, id := range idParamIns {
		ids[BlockContextIdParamIn0+SpirvId(i)] = SpirvUsedId{Id: id, Name: fmt.Sprintf("param_in_%d", i)}
	}

	// Pre-allocate SPIR-V labels ID for GCN CFG blocks.
	labelIds := make([]SpirvId, len(shader.Cfg.Blocks))
	for i := range shader.Cfg.Blocks {
		labelIds[i] = b.AllocId()
		b.EmitName(labelIds[i], fmt.Sprintf("bb_%d", i))
	}

	// Prepare block context with all GCN and our internal IDs.
	blockContext := SpirvBlockContext{
		Stage:           shader.Stage,
		Address:         shader.Address,
		LabelIds:        labelIds,
		Ids:             ids,
		ConstIds:        constIds,
		GcnSgprArrayId:  idSgprArrayVar,
		GcnVgprArrayId:  idVgprArrayVar,
		GcnSpecialIds:   gcnSpecialIds,
		GcnConstIds:     gcnConstIds,
		Resources:       resources,
		PsInputControls: ctx.PsInputControls,
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
		Address:   shader.Address,
		Stage:     shader.Stage,
		Code:      b.Assemble(),
		Resources: resources,
	}, nil
}
