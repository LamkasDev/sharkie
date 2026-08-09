package spirv

import (
	"fmt"
	"math"
	"math/bits"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"go101.org/nstd"
)

type SpirvShader struct {
	GcnShader    *GcnShader
	Code         []uint32
	StaticLayout map[*gcnSpec.Instruction]ShaderResourceBinding
}

func NewSpirvShader(shader *GcnShader, ctx SpirvShaderContext) (*SpirvShader, error) {
	staticLayout := BuildStaticLayout(AnalyzeResources(shader, ctx), shader)
	b := NewSpvBuilder()

	// Capabilities.
	b.EmitCapability(spec.SpvCapShader)
	b.EmitCapability(spec.SpvCapInt8)
	b.EmitCapability(spec.SpvCapInt16)
	b.EmitCapability(spec.SpvCapInt64)
	b.EmitCapability(spec.SpvCapSampled1D)
	b.EmitCapability(spec.SpvCapImage1D)
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
	typeV4Bool := b.EmitTypeVector(typeBool, 4)
	typeInt := b.EmitTypeInt(32, true)
	typeInt64 := b.EmitTypeInt(64, true)
	typeUint := b.EmitTypeInt(32, false)
	typeUint8 := b.EmitTypeInt(8, false)
	typeUint16 := b.EmitTypeInt(16, false)
	typeUint64 := b.EmitTypeInt(64, false)
	idFnType := b.EmitTypeFunction(typeVoid)

	typeV2Int := b.EmitTypeVector(typeInt, 2)
	typeV3Int := b.EmitTypeVector(typeInt, 3)
	typeV4Int := b.EmitTypeVector(typeInt, 4)
	typeV2Uint := b.EmitTypeVector(typeUint, 2)
	typeV3Uint := b.EmitTypeVector(typeUint, 3)
	typeV4Uint := b.EmitTypeVector(typeUint, 4)
	typeStructUintUint := b.EmitTypeStruct(typeUint, typeUint)

	typeFloat := b.EmitTypeFloat(32)
	typeV2Float := b.EmitTypeVector(typeFloat, 2)
	typeV3Float := b.EmitTypeVector(typeFloat, 3)
	typeV4Float := b.EmitTypeVector(typeFloat, 4)

	typeImage1d := b.EmitTypeImage(typeFloat, 0, 0, 0, 0, 1, 0)
	typeSampledImage1d := b.EmitTypeSampledImage(typeImage1d)
	typePtrUniformSampledImage1d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage1d)

	typeStorageImage1d := b.EmitTypeImage(typeFloat, 0, 0, 0, 0, 2, 0)
	typePtrUniformStorageImage1d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStorageImage1d)

	typeImage2d := b.EmitTypeImage(typeFloat, 1, 0, 0, 0, 1, 0)
	typeSampledImage2d := b.EmitTypeSampledImage(typeImage2d)
	typePtrUniformSampledImage2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage2d)

	typeStorageImage2d := b.EmitTypeImage(typeFloat, 1, 0, 0, 0, 2, 0)
	typePtrUniformStorageImage2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStorageImage2d)

	typeImage3d := b.EmitTypeImage(typeFloat, 2, 0, 0, 0, 1, 0)
	typeSampledImage3d := b.EmitTypeSampledImage(typeImage3d)
	typePtrUniformSampledImage3d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage3d)

	typeStorageImage3d := b.EmitTypeImage(typeFloat, 2, 0, 0, 0, 2, 0)
	typePtrUniformStorageImage3d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStorageImage3d)

	typeImage2dArray := b.EmitTypeImage(typeFloat, 1, 0, 1, 0, 1, 0)
	typeSampledImage2dArray := b.EmitTypeSampledImage(typeImage2dArray)
	typePtrUniformSampledImage2dArray := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeSampledImage2dArray)

	typeStorageImage2dArray := b.EmitTypeImage(typeFloat, 1, 0, 1, 0, 2, 0)
	typePtrUniformStorageImage2dArray := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStorageImage2dArray)

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
	typePtrPsbUint8 := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeUint8)
	typePtrPsbUint16 := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeUint16)
	typePtrPsbV2Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV2Uint)
	typePtrPsbV3Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV3Uint)
	typePtrPsbV4Uint := b.EmitTypePointer(spec.SpvStoragePhysicalStorageBuffer, typeV4Uint)
	b.EmitDecorate(typePtrPsbUint, spec.SpvDecorationArrayStride, 4)
	b.EmitDecorate(typePtrPsbUint8, spec.SpvDecorationArrayStride, 1)
	b.EmitDecorate(typePtrPsbUint16, spec.SpvDecorationArrayStride, 2)
	b.EmitDecorate(typePtrPsbV2Uint, spec.SpvDecorationArrayStride, 8)
	b.EmitDecorate(typePtrPsbV3Uint, spec.SpvDecorationArrayStride, 12)
	b.EmitDecorate(typePtrPsbV4Uint, spec.SpvDecorationArrayStride, 16)

	// Push constant struct.
	// struct StubPushConstants {
	// 	PhysicalStorageBuffer uint* UserDataAddress;
	//  uint64_t OnionMemoryBaseAddress;
	//  uint64_t GarlicMemoryBaseAddress;
	//  uint64_t GdsMemoryBaseAddress;
	//  uint_t UserSgprCount;
	//  uint_t ShaderRsrc2;
	//  uint_t VteControl;
	//  uint_t ClipControl;
	//  float_t GbHorzClipAdj;
	//  float_t GbVertClipAdj;
	// }
	typePc := b.EmitTypeStruct(
		typePtrPsbUint, typeUint64, typeUint64, typeUint64,
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

	b.EmitMemberDecorate(typePc, 4, spec.SpvDecorationOffset, baseOffset+32)
	b.EmitMemberDecorate(typePc, 5, spec.SpvDecorationOffset, baseOffset+36)
	b.EmitMemberDecorate(typePc, 6, spec.SpvDecorationOffset, baseOffset+40)
	b.EmitMemberDecorate(typePc, 7, spec.SpvDecorationOffset, baseOffset+44)

	b.EmitMemberDecorate(typePc, 8, spec.SpvDecorationOffset, baseOffset+48)
	b.EmitMemberDecorate(typePc, 9, spec.SpvDecorationOffset, baseOffset+52)

	// Global push-constant variable.
	typePtrPc := b.EmitTypePointer(spec.SpvStoragePushConstant, typePc)
	pcVar := b.EmitVariable(typePtrPc, spec.SpvStoragePushConstant)
	b.EmitName(pcVar, "push_constants")
	b.EmitDecorate(pcVar, spec.SpvDecorationAliasedPointer)

	idStaticCount := b.EmitConstantUint(typeUint, MaxStaticBindings)

	// 1D images.
	typeStaticSampledArray1d := b.EmitTypeArray(typeSampledImage1d, idStaticCount)
	typePtrUniformStaticSampledArray1d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticSampledArray1d)
	idStaticTextures1d := b.EmitVariable(typePtrUniformStaticSampledArray1d, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticTextures1d, "static_textures_1d")
	b.EmitDecorate(idStaticTextures1d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticTextures1d, spec.SpvDecorationBinding, ImageBindingSampledImages1D)

	typeStaticStorageArray1d := b.EmitTypeArray(typeStorageImage1d, idStaticCount)
	typePtrUniformStaticStorageArray1d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticStorageArray1d)
	idStaticStorageTextures1d := b.EmitVariable(typePtrUniformStaticStorageArray1d, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticStorageTextures1d, "static_storage_textures_1d")
	b.EmitDecorate(idStaticStorageTextures1d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticStorageTextures1d, spec.SpvDecorationBinding, ImageBindingStorageImages1D)

	// 2D images.
	typeStaticSampledArray2d := b.EmitTypeArray(typeSampledImage2d, idStaticCount)
	typePtrUniformStaticSampledArray2d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticSampledArray2d)
	idStaticTextures2d := b.EmitVariable(typePtrUniformStaticSampledArray2d, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticTextures2d, "static_textures_2d")
	b.EmitDecorate(idStaticTextures2d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticTextures2d, spec.SpvDecorationBinding, ImageBindingSampledImages2D)

	typeStaticStorageArray := b.EmitTypeArray(typeStorageImage2d, idStaticCount)
	typePtrUniformStaticStorageArray := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticStorageArray)
	idStaticStorageTextures2d := b.EmitVariable(typePtrUniformStaticStorageArray, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticStorageTextures2d, "static_storage_textures_2d")
	b.EmitDecorate(idStaticStorageTextures2d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticStorageTextures2d, spec.SpvDecorationBinding, ImageBindingStorageImages2D)

	// 3D images.
	typeStaticSampledArray3d := b.EmitTypeArray(typeSampledImage3d, idStaticCount)
	typePtrUniformStaticSampledArray3d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticSampledArray3d)
	idStaticTextures3d := b.EmitVariable(typePtrUniformStaticSampledArray3d, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticTextures3d, "static_textures_3d")
	b.EmitDecorate(idStaticTextures3d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticTextures3d, spec.SpvDecorationBinding, ImageBindingSampledImages3D)

	typeStaticStorageArray3d := b.EmitTypeArray(typeStorageImage3d, idStaticCount)
	typePtrUniformStaticStorageArray3d := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticStorageArray3d)
	idStaticStorageTextures3d := b.EmitVariable(typePtrUniformStaticStorageArray3d, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticStorageTextures3d, "static_storage_textures_3d")
	b.EmitDecorate(idStaticStorageTextures3d, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticStorageTextures3d, spec.SpvDecorationBinding, ImageBindingStorageImages3D)

	// 2D arrays.
	typeStaticSampledArray2dArray := b.EmitTypeArray(typeSampledImage2dArray, idStaticCount)
	typePtrUniformStaticSampledArray2dArray := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticSampledArray2dArray)
	idStaticTextures2dArray := b.EmitVariable(typePtrUniformStaticSampledArray2dArray, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticTextures2dArray, "static_textures_2d_array")
	b.EmitDecorate(idStaticTextures2dArray, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticTextures2dArray, spec.SpvDecorationBinding, ImageBindingSampledImages2DArray)

	typeStaticStorageArray2dArray := b.EmitTypeArray(typeStorageImage2dArray, idStaticCount)
	typePtrUniformStaticStorageArray2dArray := b.EmitTypePointer(spec.SpvStorageUniformConstant, typeStaticStorageArray2dArray)
	idStaticStorageTextures2dArray := b.EmitVariable(typePtrUniformStaticStorageArray2dArray, spec.SpvStorageUniformConstant)
	b.EmitName(idStaticStorageTextures2dArray, "static_storage_textures_2d_array")
	b.EmitDecorate(idStaticStorageTextures2dArray, spec.SpvDecorationDescriptorSet, DescriptorSetSlotImages)
	b.EmitDecorate(idStaticStorageTextures2dArray, spec.SpvDecorationBinding, ImageBindingStorageImages2DArray)

	// Address translation.
	typeStructAddressTranslationEntry := b.EmitTypeStruct(typeUint64, typeUint64, typeUint64, typeUint64)
	b.EmitMemberDecorate(typeStructAddressTranslationEntry, 0, spec.SpvDecorationOffset, 0)
	b.EmitMemberDecorate(typeStructAddressTranslationEntry, 1, spec.SpvDecorationOffset, 8)
	b.EmitMemberDecorate(typeStructAddressTranslationEntry, 2, spec.SpvDecorationOffset, 16)
	b.EmitMemberDecorate(typeStructAddressTranslationEntry, 3, spec.SpvDecorationOffset, 24)

	idAddressTranslationCount := b.EmitConstantUint(typeUint, AddressTranslationCount)
	typeAddressTranslationArray := b.EmitTypeArray(typeStructAddressTranslationEntry, idAddressTranslationCount)
	b.EmitDecorate(typeAddressTranslationArray, spec.SpvDecorationArrayStride, 32)

	typeAddressTranslationBuffer := b.EmitTypeStruct(typeAddressTranslationArray)
	b.EmitDecorate(typeAddressTranslationBuffer, spec.SpvDecorationBlock)
	b.EmitMemberDecorate(typeAddressTranslationBuffer, 0, spec.SpvDecorationOffset, 0)

	typePtrStorageAddressTranslationBuffer := b.EmitTypePointer(spec.SpvStorageStorageBuffer, typeAddressTranslationBuffer)
	typeAddressTranslationBufferVar := b.EmitVariable(typePtrStorageAddressTranslationBuffer, spec.SpvStorageStorageBuffer)
	b.EmitName(typeAddressTranslationBufferVar, "address_translation_buffer")
	b.EmitDecorate(typeAddressTranslationBufferVar, spec.SpvDecorationDescriptorSet, DescriptorSetSlotGlobal)
	b.EmitDecorate(typeAddressTranslationBufferVar, spec.SpvDecorationBinding, GlobalBindingAddressTranslation)
	b.EmitDecorate(typeAddressTranslationBufferVar, spec.SpvDecorationNonWritable)

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
	var idFrontFacing SpirvId
	switch shader.Stage {
	case GcnShaderStageVertex:
		ctxVs := ctx.(SpirvVertexShaderContext)
		typePosOut = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
		b.EmitName(typePosOut, "pos_out")
		b.EmitDecorate(typePosOut, spec.SpvDecorationBuiltIn, spec.SpvBuiltInPosition)

		interfaceIds = append(interfaceIds, typePosOut)

		if ctxVs.ClipDistEnable != 0 {
			count := uint32(bits.OnesCount8(ctxVs.ClipDistEnable))
			idCount := b.EmitConstantUint(typeUint, count)
			typeArray := b.EmitTypeArray(typeFloat, idCount)
			typePtrArray := b.EmitTypePointer(spec.SpvStorageOutput, typeArray)
			idClipDist := b.EmitVariable(typePtrArray, spec.SpvStorageOutput)
			b.EmitName(idClipDist, "clip_distance")
			b.EmitDecorate(idClipDist, spec.SpvDecorationBuiltIn, spec.SpvBuiltInClipDistance)
			interfaceIds = append(interfaceIds, idClipDist)
		}

		if ctxVs.CullDistEnable != 0 {
			count := uint32(bits.OnesCount8(ctxVs.CullDistEnable))
			idCount := b.EmitConstantUint(typeUint, count)
			typeArray := b.EmitTypeArray(typeFloat, idCount)
			typePtrArray := b.EmitTypePointer(spec.SpvStorageOutput, typeArray)
			idCullDist := b.EmitVariable(typePtrArray, spec.SpvStorageOutput)
			b.EmitName(idCullDist, "cull_distance")
			b.EmitDecorate(idCullDist, spec.SpvDecorationBuiltIn, spec.SpvBuiltInCullDistance)
			interfaceIds = append(interfaceIds, idCullDist)
		}

		// Use VS_EXPORT_COUNT to only emit required output parameters.
		exportCount := ctxVs.VsExportCount + 1
		for i := uint32(0); i < exportCount; i++ {
			idParamOuts[i] = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
			b.EmitDecorate(idParamOuts[i], spec.SpvDecorationLocation, i)
			interfaceIds = append(interfaceIds, idParamOuts[i])
		}
	case GcnShaderStageFragment:
		ctxFs := ctx.(SpirvFragmentShaderContext)
		for i := range idColorOuts {
			idColorOuts[i] = b.EmitVariable(idPtrOutV4, spec.SpvStorageOutput)
			b.EmitDecorate(idColorOuts[i], spec.SpvDecorationLocation, uint32(i))
			interfaceIds = append(interfaceIds, idColorOuts[i])
		}

		// Use PsInControl.NUM_INTERP to determine which parameters to declare.
		// Bits 5:0 specify the number of parameters to interpolate (0-32).
		numInterp := ctxFs.PsInControl & 0x3F
		var usedParamTypes []uint8
		for i := uint32(0); i < numInterp; i++ {
			control := ctxFs.PsInputControls[i]
			if (control & 0x3F) != 0x20 { // 0x20 means no match
				usedParamTypes = append(usedParamTypes, uint8(i))
			}
		}

		for _, paramIdx := range usedParamTypes {
			control := ctxFs.PsInputControls[paramIdx]
			offset := control & 0x3F
			if match := offset&0x20 == 0; match {
				// Vertex shader match found.
				location := offset & 0x1F
				idParamIns[paramIdx] = b.EmitVariable(b.EmitTypePointer(spec.SpvStorageInput, typeV4Float), spec.SpvStorageInput)
				b.EmitName(idParamIns[paramIdx], fmt.Sprintf("param_in_%d", paramIdx))
				b.EmitDecorate(idParamIns[paramIdx], spec.SpvDecorationLocation, location)
				if flat := (control>>10)&1 == 1; flat {
					b.EmitDecorate(idParamIns[paramIdx], spec.SpvDecorationFlat)
				}
				interfaceIds = append(interfaceIds, idParamIns[paramIdx])
			}
		}

		// Handle system inputs based on address bits.
		if (ctxFs.PsInputAddress>>8)&0xF != 0 {
			typePtrInputV4F := b.EmitTypePointer(spec.SpvStorageInput, typeV4Float)
			typeFragCoord := b.EmitVariable(typePtrInputV4F, spec.SpvStorageInput)
			b.EmitName(typeFragCoord, "frag_coord")
			b.EmitDecorate(typeFragCoord, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFragCoord)
			interfaceIds = append(interfaceIds, typeFragCoord)
		}

		if ctxFs.FrontFaceEnable {
			typePtrInputBool := b.EmitTypePointer(spec.SpvStorageInput, typeBool)
			idFrontFacing = b.EmitVariable(typePtrInputBool, spec.SpvStorageInput)
			b.EmitName(idFrontFacing, "front_facing")
			b.EmitDecorate(idFrontFacing, spec.SpvDecorationBuiltIn, spec.SpvBuiltInFrontFacing)
			interfaceIds = append(interfaceIds, idFrontFacing)
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
		ctxFs := ctx.(SpirvFragmentShaderContext)
		b.EmitExecutionMode(idMain, spec.SpvExecModeOriginUpperLeft)
		if ctxFs.DepthBeforeShader || ctxFs.ZOrder == 1 || ctxFs.ZOrder == 3 {
			b.EmitExecutionMode(idMain, spec.SpvExecModeEarlyFragmentTests)
		}
		if typeFragDepthOut != 0 {
			b.EmitExecutionMode(idMain, spec.SpvExecModeDepthReplacing)
		}
	case GcnShaderStageCompute:
		ctxCs := ctx.(SpirvComputeShaderContext)
		b.EmitExecutionMode(idMain, spec.SpvExecModeLocalSize, nstd.Clamp(ctxCs.ThreadX, 0, 1024), ctxCs.ThreadY, ctxCs.ThreadZ)
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
	idNullSgprArray := b.EmitConstantNull(typeSgprArray)
	idNullVgprArray := b.EmitConstantNull(typeVgprArray)
	b.EmitDeferredLocalVariableInit(typePtrSgprArray, idSgprArrayVar, idNullSgprArray)
	b.EmitDeferredLocalVariableInit(typePtrVgprArray, idVgprArrayVar, idNullVgprArray)

	// Emit variable tracking which pixels are valid.
	var idIsValidPixel SpirvId
	if shader.Stage == GcnShaderStageFragment {
		idIsValidPixel = b.AllocId()
		typePtrFnBool := b.EmitTypePointer(spec.SpvStorageFunction, typeBool)
		b.EmitDeferredLocalVariable(typePtrFnBool, idIsValidPixel)
	}

	// Emit LDS scratchpad variable (32kB = 8192 Dwords).
	var idLdsArray SpirvId
	if shader.Stage == GcnShaderStageCompute {
		typeLdsArray := b.EmitTypeArray(typeUint, b.EmitConstantUint(typeUint, 8192))
		typePtrLdsArray := b.EmitTypePointer(spec.SpvStorageWorkgroup, typeLdsArray)
		idLdsArray = b.EmitVariable(typePtrLdsArray, spec.SpvStorageWorkgroup)
		b.EmitName(idLdsArray, "lds_array")
	}

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
	constIds[ConstIdFloat4] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(4.0), Name: "4.0"}
	constIds[ConstIdFloat255] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(255.0), Name: "255.0"}
	constIds[ConstIdFloat65535] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(65535.0), Name: "65535.0"}
	constIds[ConstIdFloatMin] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(-math.MaxFloat32), Name: "-MaxFloat32"}
	constIds[ConstIdFloatMax] = SpirvUsedId{Id: b.AllocId(), Value: math.Float32bits(math.MaxFloat32), Name: "MaxFloat32"}

	// Prepare internal IDs.
	ids := map[SpirvId]SpirvUsedId{
		// Constants.
		BlockContextIdFalse:           {Id: idFalse, Name: "false"},
		BlockContextIdTrue:            {Id: idTrue, Name: "true"},
		BlockContextIdTypeZeroVec4:    {Id: typeZeroVec4, Name: "zero_vec4_t"},
		BlockContextIdTypeGlsl:        {Id: typeGLSL, Name: "glsl_t"},
		BlockContextIdTypeDebugPrintf: {Id: typeDebugPrintf, Name: "debug_printf_t"},

		// Scalar types.
		BlockContextIdTypeVoid:   {Id: typeVoid, Name: "void_t"},
		BlockContextIdTypeBool:   {Id: typeBool, Name: "bool_t"},
		BlockContextIdTypeV4Bool: {Id: typeV4Bool, Name: "v4bool_t"},
		BlockContextIdTypeFloat:  {Id: typeFloat, Name: "float_t"},
		BlockContextIdTypeInt:    {Id: typeInt, Name: "int_t"},
		BlockContextIdTypeUint:   {Id: typeUint, Name: "uint_t"},
		BlockContextIdTypeUint8:  {Id: typeUint8, Name: "uint8_t"},
		BlockContextIdTypeUint16: {Id: typeUint16, Name: "uint16_t"},
		BlockContextIdTypeUint64: {Id: typeUint64, Name: "uint64_t"},
		BlockContextIdTypeInt64:  {Id: typeInt64, Name: "int64_t"},

		// Vector & composite types.
		BlockContextIdTypeV2Float:        {Id: typeV2Float, Name: "v2float_t"},
		BlockContextIdTypeV3Float:        {Id: typeV3Float, Name: "v3float_t"},
		BlockContextIdTypeV4Float:        {Id: typeV4Float, Name: "v4float_t"},
		BlockContextIdTypeV2Int:          {Id: typeV2Int, Name: "v2int_t"},
		BlockContextIdTypeV3Int:          {Id: typeV3Int, Name: "v3int_t"},
		BlockContextIdTypeV4Int:          {Id: typeV4Int, Name: "v4int_t"},
		BlockContextIdTypeV2Uint:         {Id: typeV2Uint, Name: "v2uint_t"},
		BlockContextIdTypeV3Uint:         {Id: typeV3Uint, Name: "v3uint_t"},
		BlockContextIdTypeV4Uint:         {Id: typeV4Uint, Name: "v4uint_t"},
		BlockContextIdTypeStructUintUint: {Id: typeStructUintUint, Name: "struct_uint_uint_t"},

		// 1D images.
		BlockContextIdTypeImage1d:              {Id: typeImage1d, Name: "image_1d_t"},
		BlockContextIdTypeSampledImage1d:       {Id: typeSampledImage1d, Name: "sampled_image_1d_t"},
		BlockContextIdPtrUniformSampledImage1d: {Id: typePtrUniformSampledImage1d, Name: "ptr_uniform_sampled_image_1d_t"},
		BlockContextIdStaticTextures1d:         {Id: idStaticTextures1d, Name: "static_textures_1d"},
		BlockContextIdTypeStorageImage1d:       {Id: typeStorageImage1d, Name: "storage_image_1d_t"},
		BlockContextIdPtrUniformStorageImage1d: {Id: typePtrUniformStorageImage1d, Name: "ptr_uniform_storage_image_1d_t"},
		BlockContextIdStaticStorageTextures1d:  {Id: idStaticStorageTextures1d, Name: "static_storage_textures_1d"},

		// 2D images.
		BlockContextIdTypeImage2d:              {Id: typeImage2d, Name: "image_2d_t"},
		BlockContextIdTypeSampledImage2d:       {Id: typeSampledImage2d, Name: "sampled_image_2d_t"},
		BlockContextIdPtrUniformSampledImage2d: {Id: typePtrUniformSampledImage2d, Name: "ptr_uniform_sampled_image_2d_t"},
		BlockContextIdStaticTextures2d:         {Id: idStaticTextures2d, Name: "static_textures_2d"},
		BlockContextIdTypeStorageImage2d:       {Id: typeStorageImage2d, Name: "storage_image_2d_t"},
		BlockContextIdPtrUniformStorageImage2d: {Id: typePtrUniformStorageImage2d, Name: "ptr_uniform_storage_image_2d_t"},
		BlockContextIdStaticStorageTextures2d:  {Id: idStaticStorageTextures2d, Name: "static_storage_textures_2d"},

		// 3D images.
		BlockContextIdTypeImage3d:              {Id: typeImage3d, Name: "image_3d_t"},
		BlockContextIdTypeSampledImage3d:       {Id: typeSampledImage3d, Name: "sampled_image_3d_t"},
		BlockContextIdPtrUniformSampledImage3d: {Id: typePtrUniformSampledImage3d, Name: "ptr_uniform_sampled_image_3d_t"},
		BlockContextIdStaticTextures3d:         {Id: idStaticTextures3d, Name: "static_textures_3d"},
		BlockContextIdTypeStorageImage3d:       {Id: typeStorageImage3d, Name: "storage_image_3d_t"},
		BlockContextIdPtrUniformStorageImage3d: {Id: typePtrUniformStorageImage3d, Name: "ptr_uniform_storage_image_3d_t"},
		BlockContextIdStaticStorageTextures3d:  {Id: idStaticStorageTextures3d, Name: "static_storage_textures_3d"},

		// 2D arrays.
		BlockContextIdTypeImage2dArray:              {Id: typeImage2dArray, Name: "image_2d_array_t"},
		BlockContextIdTypeSampledImage2dArray:       {Id: typeSampledImage2dArray, Name: "sampled_image_2d_array_t"},
		BlockContextIdPtrUniformSampledImage2dArray: {Id: typePtrUniformSampledImage2dArray, Name: "ptr_uniform_sampled_image_2d_array_t"},
		BlockContextIdStaticTextures2dArray:         {Id: idStaticTextures2dArray, Name: "static_textures_2d_array"},
		BlockContextIdTypeStorageImage2dArray:       {Id: typeStorageImage2dArray, Name: "storage_image_2d_array_t"},
		BlockContextIdPtrUniformStorageImage2dArray: {Id: typePtrUniformStorageImage2dArray, Name: "ptr_uniform_storage_image_2d_array_t"},
		BlockContextIdStaticStorageTextures2dArray:  {Id: idStaticStorageTextures2dArray, Name: "static_storage_textures_2d_array"},

		// Pointers.
		BlockContextIdPtrPcPsbUint: {Id: typePtrPcPsbUint, Name: "ptr_pc_psb_uint_t"},
		BlockContextIdPtrPcUint64:  {Id: typePtrPcUint64, Name: "ptr_pc_uint64_t"},
		BlockContextIdPtrPcUint:    {Id: typePtrPcUint, Name: "ptr_pc_uint_t"},
		BlockContextIdPtrPcFloat:   {Id: typePtrPcFloat, Name: "ptr_pc_float_t"},
		BlockContextIdPtrPsbUint:   {Id: typePtrPsbUint, Name: "ptr_psb_uint_t"},
		BlockContextIdPtrPsbUint8:  {Id: typePtrPsbUint8, Name: "ptr_psb_uint8_t"},
		BlockContextIdPtrPsbUint16: {Id: typePtrPsbUint16, Name: "ptr_psb_uint16_t"},
		BlockContextIdPtrPsbV2Uint: {Id: typePtrPsbV2Uint, Name: "ptr_psb_v2_uint_t"},
		BlockContextIdPtrPsbV3Uint: {Id: typePtrPsbV3Uint, Name: "ptr_psb_v3_uint_t"},
		BlockContextIdPtrPsbV4Uint: {Id: typePtrPsbV4Uint, Name: "ptr_psb_v4_uint_t"},
		BlockContextIdPtrFnUint:    {Id: typePtrFnUint, Name: "ptr_fn_uint_t"},

		// System variables & built-ins.
		BlockContextIdPcVar:                     {Id: pcVar, Name: "pc_var"},
		BlockContextIdAddressTranslationBuffer:  {Id: typeAddressTranslationBufferVar, Name: "address_translation_buffer_var_t"},
		BlockContextIdSubgroupLocalInvocationId: {Id: typeSubgroupLocalInvocationId, Name: "subgroup_local_invocation_id_t"},
		BlockContextIdVertexIndex:               {Id: typeVertexIndex, Name: "vertex_index_t"},
		BlockContextIdInstanceIndex:             {Id: typeInstanceIndex, Name: "instance_index_t"},
		BlockContextIdWorkgroupId:               {Id: idWorkgroupId, Name: "workgroup_id"},
		BlockContextIdLocalInvocationId:         {Id: idLocalInvocationId, Name: "local_invocation_id"},
		BlockContextIdIsValidPixel:              {Id: idIsValidPixel, Name: "is_valid_pixel"},
		BlockContextIdLdsArray:                  {Id: idLdsArray, Name: "lds_array"},

		// Pipeline outputs.
		BlockContextIdPosOut:       {Id: typePosOut, Name: "pos_out_t"},
		BlockContextIdFragDepthOut: {Id: typeFragDepthOut, Name: "frag_depth_out_t"},
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

	// Pre-allocate continue block IDs for loop headers.
	for i, block := range shader.Cfg.Blocks {
		if block.IsLoopHeader && block.ContinueBlockId == block.Id {
			ids[BlockContextIdContinueBlocks+SpirvId(block.Id)] = SpirvUsedId{Id: b.AllocId(), Name: fmt.Sprintf("continue_%d", i)}
		}
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
		Address:        shader.Address,
		LabelIds:       labelIds,
		Ids:            ids,
		ConstIds:       constIds,
		GcnSgprArrayId: idSgprArrayVar,
		GcnVgprArrayId: idVgprArrayVar,
		GcnSpecialIds:  gcnSpecialIds,
		GcnConstIds:    gcnConstIds,
		StaticLayout:   staticLayout,
		Context:        ctx,
	}

	// Function body.
	b.EmitFunction(typeVoid, idMain, spec.SpvFunctionControlNone, idFnType)

	// Emit reachable blocks in reverse post-order (entry block first).
	rpoBlockIds := shader.Cfg.ReversePostOrder()
	emittedBlockIds := make([]bool, len(shader.Cfg.Blocks))

	// Emit reachable blocks.
	for _, blockId := range rpoBlockIds {
		emitBlock(b, &shader.Cfg, &shader.Cfg.Blocks[blockId], &blockContext)
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
	idNullFnUint := b.EmitConstantNull(typeUint)
	for _, c := range blockContext.GcnSpecialIds {
		if c.Id == GcnSpecIdReserved || !c.Used {
			continue // reserved.
		}
		b.EmitDeferredLocalVariableInit(typePtrFnUint, c.Id, idNullFnUint)
		b.EmitName(c.Id, c.Name)
	}

	// Andddd we're done :)
	b.EmitFunctionEnd()

	return &SpirvShader{
		GcnShader:    shader,
		Code:         b.Assemble(),
		StaticLayout: staticLayout,
	}, nil
}
