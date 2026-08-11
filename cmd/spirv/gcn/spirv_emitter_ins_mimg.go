package gcn

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

const (
	DiscoveryMapBufferEntries = 65536
	DiscoveryMapBufferSize    = DiscoveryMapBufferEntries * 4
)

func EmitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeV2Int := ctx.GetId(BlockContextIdTypeV2Int)
	typeV3Int := ctx.GetId(BlockContextIdTypeV3Int)
	typeV2Float := ctx.GetId(BlockContextIdTypeV2Float)
	typeV3Float := ctx.GetId(BlockContextIdTypeV3Float)
	typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)

	resourceType := extractResourceType(b, ctx, details)
	is1D := b.EmitIEqual(typeBool, resourceType, b.EmitConstantUint(typeUint, gcn.GcnImageTypeColor1D))
	is3D := b.EmitIEqual(typeBool, resourceType, b.EmitConstantUint(typeUint, gcn.GcnImageTypeColor3D))
	isCubeOrArray := b.EmitUGreaterThanEqual(typeBool, resourceType, b.EmitConstantUint(typeUint, gcn.GcnImageTypeCubeOrArray))

	binding, ok := ctx.StaticLayout[instr]
	if !ok {
		panic(fmt.Sprintf("static binding not found for instruction %v", instr))
	}
	bindingIndex := b.EmitConstantUint(typeUint, binding.BindingIndex)
	switch details.Op {
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz, gcnSpec.MimgOpSampleB, gcnSpec.MimgOpSampleLzO, gcnSpec.MimgOpSampleC, gcnSpec.MimgOpSampleL:
		idStaticTextures1D := ctx.GetId(BlockContextIdStaticTextures1d)
		idStaticTextures2D := ctx.GetId(BlockContextIdStaticTextures2d)
		idStaticTextures3D := ctx.GetId(BlockContextIdStaticTextures3d)
		idStaticTextures2DArray := ctx.GetId(BlockContextIdStaticTextures2dArray)

		typeSampledImage1D := ctx.GetId(BlockContextIdTypeSampledImage1d)
		typeSampledImage2D := ctx.GetId(BlockContextIdTypeSampledImage2d)
		typeSampledImage3D := ctx.GetId(BlockContextIdTypeSampledImage3d)
		typeSampledImage2DArray := ctx.GetId(BlockContextIdTypeSampledImage2dArray)

		typePtrUniformSampledImage1D := ctx.GetId(BlockContextIdPtrUniformSampledImage1d)
		typePtrUniformSampledImage2D := ctx.GetId(BlockContextIdPtrUniformSampledImage2d)
		typePtrUniformSampledImage3D := ctx.GetId(BlockContextIdPtrUniformSampledImage3d)
		typePtrUniformSampledImage2DArray := ctx.GetId(BlockContextIdPtrUniformSampledImage2dArray)

		ptr1D := b.EmitAccessChain(typePtrUniformSampledImage1D, idStaticTextures1D, bindingIndex)
		sampledImage1D := b.EmitLoad(typeSampledImage1D, ptr1D)
		ptr2D := b.EmitAccessChain(typePtrUniformSampledImage2D, idStaticTextures2D, bindingIndex)
		sampledImage2D := b.EmitLoad(typeSampledImage2D, ptr2D)
		ptr3D := b.EmitAccessChain(typePtrUniformSampledImage3D, idStaticTextures3D, bindingIndex)
		sampledImage3D := b.EmitLoad(typeSampledImage3D, ptr3D)
		ptr2DArray := b.EmitAccessChain(typePtrUniformSampledImage2DArray, idStaticTextures2DArray, bindingIndex)
		sampledImage2DArray := b.EmitLoad(typeSampledImage2DArray, ptr2DArray)

		// Flags for our supported prefixes.
		isOffset := details.Op == gcnSpec.MimgOpSampleLzO
		isBias := details.Op == gcnSpec.MimgOpSampleB
		isLz := details.Op == gcnSpec.MimgOpSampleLz || details.Op == gcnSpec.MimgOpSampleLzO
		isLod := details.Op == gcnSpec.MimgOpSampleL
		isPcf := details.Op == gcnSpec.MimgOpSampleC
		vaddrOffset := uint32(0)

		// { offset } - send X, Y, Z integer offsets (packed into 1 dword) to offset XYZ address.
		var offsetX, offsetY, offsetZ SpirvId
		if isOffset {
			rawOffsetUint := ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset)
			vaddrOffset++

			// Extract 6-bit signed offsets: X=[5:0], Y=[13:8], Z=[21:16].
			rawOffsetInt := b.EmitBitcast(typeInt, rawOffsetUint)
			offsetX = b.EmitBitFieldSExtract(typeInt, rawOffsetInt, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 6))
			offsetY = b.EmitBitFieldSExtract(typeInt, rawOffsetInt, b.EmitConstantUint(typeUint, 8), b.EmitConstantUint(typeUint, 6))
			offsetZ = b.EmitBitFieldSExtract(typeInt, rawOffsetInt, b.EmitConstantUint(typeUint, 16), b.EmitConstantUint(typeUint, 6))
		}

		// { bias } - add this BIAS to the LOD TA computes.
		var bias SpirvId
		if isBias {
			bias = b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset))
			vaddrOffset++
		}

		// { z-compare } - PCF.
		var zCompare SpirvId
		if isPcf {
			zCompare = b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset))
			vaddrOffset++
		}

		// { body } - coordinates from remaining VGPRs.
		coordX := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset))
		coordY := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset+1))
		coordZ := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset+2))

		coord1D := coordX
		coord2D := b.EmitCompositeConstruct(typeV2Float, coordX, coordY)
		coord3D := b.EmitCompositeConstruct(typeV3Float, coordX, coordY, coordZ)
		coord2DArray := coord3D

		// Image sample.
		resVec4 := EmitResourceTypeSwitch(b, typeV4Float, is1D, is3D, isCubeOrArray,
			sampledImage1D, sampledImage2D, sampledImage3D, sampledImage2DArray,
			coord1D, coord2D, coord3D, coord2DArray,
			func(imageId, coord SpirvId, resType int) SpirvId {
				// Manually apply the offset to the coordinates before sampling.
				if isOffset {
					// Convert the extracted integer offsets into floats for math.
					var offsetFloat SpirvId
					switch resType {
					case ResType1D:
						offsetFloat = b.EmitConvertSToF(typeFloat, offsetX)
					case ResType2D, ResType2DArray:
						offsetFloat = b.EmitConvertSToF(typeV2Float, b.EmitCompositeConstruct(typeV2Int, offsetX, offsetY))
					case ResType3D:
						offsetFloat = b.EmitConvertSToF(typeV3Float, b.EmitCompositeConstruct(typeV3Int, offsetX, offsetY, offsetZ))
					}

					if details.Unorm {
						// Coordinates are already in absolute texels (e.g., 0 to 1024).
						// We can just add the float offset directly to the UVs.
						switch resType {
						case ResType1D:
							coord = b.EmitFAdd(typeFloat, coord, offsetFloat)
						case ResType2D:
							coord = b.EmitFAdd(typeV2Float, coord, offsetFloat)
						case ResType3D:
							coord = b.EmitFAdd(typeV3Float, coord, offsetFloat)
						case ResType2DArray:
							// coord is V3 (U, V, Layer). We only add the offset to U and V.
							coordU := b.EmitCompositeExtract(typeFloat, coord, 0)
							coordV := b.EmitCompositeExtract(typeFloat, coord, 1)
							coordLayer := b.EmitCompositeExtract(typeFloat, coord, 2)

							coordUV := b.EmitCompositeConstruct(typeV2Float, coordU, coordV)
							coordUV = b.EmitFAdd(typeV2Float, coordUV, offsetFloat)

							coord = b.EmitCompositeConstruct(typeV3Float,
								b.EmitCompositeExtract(typeFloat, coordUV, 0),
								b.EmitCompositeExtract(typeFloat, coordUV, 1),
								coordLayer)
						}
					} else {
						// Coordinates are 0.0 to 1.0. We must divide the texel offset by the
						// texture dimensions at the evaluated LOD to convert it into normalized space.
						var lodInt SpirvId
						if isLz {
							// If it's an explicit LOD-zero instruction, we don't need to query LOD.
							lodInt = ctx.GetConstId(ConstIdUint0)
						} else {
							// Prepare coordinate for OpImageQueryLod (strips array layer for 2DArray if needed).
							queryCoord := coord
							if resType == ResType2DArray {
								coordU := b.EmitCompositeExtract(typeFloat, coord, 0)
								coordV := b.EmitCompositeExtract(typeFloat, coord, 1)
								queryCoord = b.EmitCompositeConstruct(typeV2Float, coordU, coordV)
							}

							// OpImageQueryLod returns a vec2 where Y (component 1) is the implicit LOD.
							lodResult := b.EmitImageQueryLod(typeV2Float, imageId, queryCoord)
							implicitLodFloat := b.EmitCompositeExtract(typeFloat, lodResult, 1)

							// Cast the float LOD to an integer LOD for size querying.
							lodInt = b.EmitConvertFToS(typeInt, implicitLodFloat)
						}

						// Query texture dimensions at the computed LOD and normalize the offset.
						switch resType {
						case ResType1D:
							img := b.EmitImage(ctx.GetId(BlockContextIdTypeImage1d), imageId)
							sizeInt := b.EmitImageQuerySizeLod(typeInt, img, lodInt)
							sizeFloat := b.EmitConvertSToF(typeFloat, sizeInt)

							normOffset := b.EmitFDiv(typeFloat, offsetFloat, sizeFloat)
							coord = b.EmitFAdd(typeFloat, coord, normOffset)

						case ResType2D:
							img := b.EmitImage(ctx.GetId(BlockContextIdTypeImage2d), imageId)
							sizeInt := b.EmitImageQuerySizeLod(typeV2Int, img, lodInt)
							sizeFloat := b.EmitConvertSToF(typeV2Float, sizeInt)

							normOffset := b.EmitFDiv(typeV2Float, offsetFloat, sizeFloat)
							coord = b.EmitFAdd(typeV2Float, coord, normOffset)

						case ResType3D:
							img := b.EmitImage(ctx.GetId(BlockContextIdTypeImage3d), imageId)
							sizeInt := b.EmitImageQuerySizeLod(typeV3Int, img, lodInt)
							sizeFloat := b.EmitConvertSToF(typeV3Float, sizeInt)

							normOffset := b.EmitFDiv(typeV3Float, offsetFloat, sizeFloat)
							coord = b.EmitFAdd(typeV3Float, coord, normOffset)

						case ResType2DArray:
							img := b.EmitImage(ctx.GetId(BlockContextIdTypeImage2dArray), imageId)
							sizeInt := b.EmitImageQuerySizeLod(typeV3Int, img, lodInt) // Returns (W, H, Layers)

							// Extract only W and H for the offset normalization.
							sizeV2Int := b.EmitVectorShuffle(typeV2Int, sizeInt, sizeInt, 0, 1)
							sizeFloat := b.EmitConvertSToF(typeV2Float, sizeV2Int)
							normOffset := b.EmitFDiv(typeV2Float, offsetFloat, sizeFloat)

							// Add normalized offset to U and V, preserve Layer.
							coordU := b.EmitCompositeExtract(typeFloat, coord, 0)
							coordV := b.EmitCompositeExtract(typeFloat, coord, 1)
							coordLayer := b.EmitCompositeExtract(typeFloat, coord, 2)

							coordUV := b.EmitCompositeConstruct(typeV2Float, coordU, coordV)
							coordUV = b.EmitFAdd(typeV2Float, coordUV, normOffset)

							coord = b.EmitCompositeConstruct(typeV3Float,
								b.EmitCompositeExtract(typeFloat, coordUV, 0),
								b.EmitCompositeExtract(typeFloat, coordUV, 1),
								coordLayer)
						}
					}
				}

				// Build SPIR-V operands.
				mask := uint32(0)
				var args []SpirvId
				if isBias {
					mask |= spec.SpvImageOperandsBiasMask
					args = append(args, bias)
				}
				if isLz {
					mask |= spec.SpvImageOperandsLodMask
					args = append(args, ctx.GetConstId(ConstIdFloat0))
				}
				if isLod {
					numCoords := uint32(0)
					switch resType {
					case ResType1D:
						numCoords = 1
					case ResType2D:
						numCoords = 2
					case ResType3D, ResType2DArray:
						numCoords = 3
					}
					userLod := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+vaddrOffset+numCoords))
					mask |= spec.SpvImageOperandsLodMask
					args = append(args, userLod)
				}

				if isPcf {
					if resType == ResType3D {
						// Vulkan forbids OpImageSampleDref* on 3D images.
						// Since PCF on 3D images is technically invalid/undefined, we just return 0 to bypass validation errors.
						zero := ctx.GetConstId(ConstIdFloat0)
						return b.EmitCompositeConstruct(typeV4Float, zero, zero, zero, zero)
					}

					var sampledVal SpirvId
					if !isLz && !isLod && ctx.Stage == gcn.GcnShaderStageFragment {
						if mask == 0 {
							sampledVal = b.EmitImageSampleDrefImplicitLod(typeFloat, imageId, coord, zCompare)
						} else {
							sampledVal = b.EmitImageSampleDrefImplicitLodOperands(typeFloat, imageId, coord, zCompare, mask, args...)
						}
					} else {
						if mask == 0 {
							sampledVal = b.EmitImageSampleDrefExplicitLod(typeFloat, imageId, coord, zCompare, ctx.GetConstId(ConstIdFloat0))
						} else {
							sampledVal = b.EmitImageSampleDrefExplicitLodOperands(typeFloat, imageId, coord, zCompare, mask, args...)
						}
					}

					// Replicate scalar depth comparison result to V4Float (R, R, R, 1.0).
					return b.EmitCompositeConstruct(typeV4Float, sampledVal, sampledVal, sampledVal, ctx.GetConstId(ConstIdFloat1))
				}

				// If no explicit lod is given; rely on ImplicitLod (only valid in Fragment stages).
				if !isLz && !isLod && ctx.Stage == gcn.GcnShaderStageFragment {
					if mask == 0 {
						return b.EmitImageSampleImplicitLod(typeV4Float, imageId, coord)
					}
					return b.EmitImageSampleImplicitLodOperands(typeV4Float, imageId, coord, mask, args...)
				}

				// Explicit lod path (Compute/Vertex stages or explicit _LZ).
				if mask == 0 {
					// Fallback just in case we didn't have _LZ or bias.
					return b.EmitImageSampleExplicitLod(typeV4Float, imageId, coord, ctx.GetConstId(ConstIdFloat0))
				}

				return b.EmitImageSampleExplicitLodOperands(typeV4Float, imageId, coord, mask, args...)
			},
		)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	case gcnSpec.MimgOpLoad, gcnSpec.MimgOpLoadMip:
		idStaticTextures1D := ctx.GetId(BlockContextIdStaticTextures1d)
		idStaticTextures2D := ctx.GetId(BlockContextIdStaticTextures2d)
		idStaticTextures3D := ctx.GetId(BlockContextIdStaticTextures3d)
		idStaticTextures2DArray := ctx.GetId(BlockContextIdStaticTextures2dArray)

		typeSampledImage1D := ctx.GetId(BlockContextIdTypeSampledImage1d)
		typeSampledImage2D := ctx.GetId(BlockContextIdTypeSampledImage2d)
		typeSampledImage3D := ctx.GetId(BlockContextIdTypeSampledImage3d)
		typeSampledImage2DArray := ctx.GetId(BlockContextIdTypeSampledImage2dArray)

		typePtrUniformSampledImage1D := ctx.GetId(BlockContextIdPtrUniformSampledImage1d)
		typePtrUniformSampledImage2D := ctx.GetId(BlockContextIdPtrUniformSampledImage2d)
		typePtrUniformSampledImage3D := ctx.GetId(BlockContextIdPtrUniformSampledImage3d)
		typePtrUniformSampledImage2DArray := ctx.GetId(BlockContextIdPtrUniformSampledImage2dArray)

		typeImage1D := ctx.GetId(BlockContextIdTypeImage1d)
		typeImage2D := ctx.GetId(BlockContextIdTypeImage2d)
		typeImage3D := ctx.GetId(BlockContextIdTypeImage3d)
		typeImage2DArray := ctx.GetId(BlockContextIdTypeImage2dArray)

		ptr1D := b.EmitAccessChain(typePtrUniformSampledImage1D, idStaticTextures1D, bindingIndex)
		image1D := b.EmitImage(typeImage1D, b.EmitLoad(typeSampledImage1D, ptr1D))

		ptr2D := b.EmitAccessChain(typePtrUniformSampledImage2D, idStaticTextures2D, bindingIndex)
		image2D := b.EmitImage(typeImage2D, b.EmitLoad(typeSampledImage2D, ptr2D))

		ptr3D := b.EmitAccessChain(typePtrUniformSampledImage3D, idStaticTextures3D, bindingIndex)
		image3D := b.EmitImage(typeImage3D, b.EmitLoad(typeSampledImage3D, ptr3D))

		ptr2DArray := b.EmitAccessChain(typePtrUniformSampledImage2DArray, idStaticTextures2DArray, bindingIndex)
		image2DArray := b.EmitImage(typeImage2DArray, b.EmitLoad(typeSampledImage2DArray, ptr2DArray))

		// Coordinates from VGPRs.
		coordX := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordZ := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+2))
		coordW := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+3))

		coord1D := coordX
		coord2D := b.EmitCompositeConstruct(typeV2Int, coordX, coordY)
		coord3D := b.EmitCompositeConstruct(typeV3Int, coordX, coordY, coordZ)
		coord2DArray := coord3D

		// Image fetch.
		resVec4 := EmitResourceTypeSwitch(b, typeV4Float, is1D, is3D, isCubeOrArray,
			image1D, image2D, image3D, image2DArray,
			coord1D, coord2D, coord3D, coord2DArray,
			func(imageId, coord SpirvId, resType int) SpirvId {
				if details.Op == gcnSpec.MimgOpLoad {
					lod := b.EmitBitcast(typeInt, ctx.GetConstId(ConstIdUint0))
					return b.EmitImageFetch(typeV4Float, imageId, coord, spec.SpvImageOperandsLodMask, lod)
				}

				var lod SpirvId
				switch resType {
				case ResType1D:
					lod = coordY
				case ResType2D:
					lod = coordZ
				default: // 3D and 2DArray.
					lod = coordW
				}
				return b.EmitImageFetch(typeV4Float, imageId, coord, spec.SpvImageOperandsLodMask, lod)
			},
		)
		// ctx.EmitDebugPrintfLane(b, 0, "MimgOpLoadMip vk_idx=%d coord=%v2d lod=%d res=%v4f\n", vulkanIndex, coordinates, lod, resVec4)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	case gcnSpec.MimgOpStore:
		idStaticStorageTextures1D := ctx.GetId(BlockContextIdStaticStorageTextures1d)
		idStaticStorageTextures2D := ctx.GetId(BlockContextIdStaticStorageTextures2d)
		idStaticStorageTextures3D := ctx.GetId(BlockContextIdStaticStorageTextures3d)
		idStaticStorageTextures2DArray := ctx.GetId(BlockContextIdStaticStorageTextures2dArray)

		typeStorageImage1D := ctx.GetId(BlockContextIdTypeStorageImage1d)
		typeStorageImage2D := ctx.GetId(BlockContextIdTypeStorageImage2d)
		typeStorageImage3D := ctx.GetId(BlockContextIdTypeStorageImage3d)
		typeStorageImage2DArray := ctx.GetId(BlockContextIdTypeStorageImage2dArray)

		typePtrUniformStorageImage1D := ctx.GetId(BlockContextIdPtrUniformStorageImage1d)
		typePtrUniformStorageImage2D := ctx.GetId(BlockContextIdPtrUniformStorageImage2d)
		typePtrUniformStorageImage3D := ctx.GetId(BlockContextIdPtrUniformStorageImage3d)
		typePtrUniformStorageImage2DArray := ctx.GetId(BlockContextIdPtrUniformStorageImage2dArray)

		ptr1D := b.EmitAccessChain(typePtrUniformStorageImage1D, idStaticStorageTextures1D, bindingIndex)
		storageImage1D := b.EmitLoad(typeStorageImage1D, ptr1D)
		ptr2D := b.EmitAccessChain(typePtrUniformStorageImage2D, idStaticStorageTextures2D, bindingIndex)
		storageImage2D := b.EmitLoad(typeStorageImage2D, ptr2D)
		ptr3D := b.EmitAccessChain(typePtrUniformStorageImage3D, idStaticStorageTextures3D, bindingIndex)
		storageImage3D := b.EmitLoad(typeStorageImage3D, ptr3D)
		ptr2DArray := b.EmitAccessChain(typePtrUniformStorageImage2DArray, idStaticStorageTextures2DArray, bindingIndex)
		storageImage2DArray := b.EmitLoad(typeStorageImage2DArray, ptr2DArray)

		// Coordinates from VGPRs (X, Y).
		coordX := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordZ := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+2))

		coord1D := coordX
		coord2D := b.EmitCompositeConstruct(typeV2Int, coordX, coordY)
		coord3D := b.EmitCompositeConstruct(typeV3Int, coordX, coordY, coordZ)
		coord2DArray := coord3D

		// Data to store from VGPRs.
		dataParts := make([]SpirvId, 4)
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				dataParts[i] = b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset))
				vgprOffset++
			} else {
				if i == 3 {
					dataParts[i] = ctx.GetConstId(ConstIdFloat1)
				} else {
					dataParts[i] = ctx.GetConstId(ConstIdFloat0)
				}
			}
		}
		texel := b.EmitCompositeConstruct(typeV4Float, dataParts...)

		// Image write.
		EmitResourceTypeSwitch(b, 0, is1D, is3D, isCubeOrArray,
			storageImage1D, storageImage2D, storageImage3D, storageImage2DArray,
			coord1D, coord2D, coord3D, coord2DArray,
			func(imageId, coord SpirvId, resType int) SpirvId {
				b.EmitImageWrite(imageId, coord, texel)
				return 0
			},
		)
	case gcnSpec.MimgOpGetLod:
		idStaticTextures1D := ctx.GetId(BlockContextIdStaticTextures1d)
		idStaticTextures2D := ctx.GetId(BlockContextIdStaticTextures2d)
		idStaticTextures3D := ctx.GetId(BlockContextIdStaticTextures3d)
		idStaticTextures2DArray := ctx.GetId(BlockContextIdStaticTextures2dArray)

		typeSampledImage1D := ctx.GetId(BlockContextIdTypeSampledImage1d)
		typeSampledImage2D := ctx.GetId(BlockContextIdTypeSampledImage2d)
		typeSampledImage3D := ctx.GetId(BlockContextIdTypeSampledImage3d)
		typeSampledImage2DArray := ctx.GetId(BlockContextIdTypeSampledImage2dArray)

		typePtrUniformSampledImage1D := ctx.GetId(BlockContextIdPtrUniformSampledImage1d)
		typePtrUniformSampledImage2D := ctx.GetId(BlockContextIdPtrUniformSampledImage2d)
		typePtrUniformSampledImage3D := ctx.GetId(BlockContextIdPtrUniformSampledImage3d)
		typePtrUniformSampledImage2DArray := ctx.GetId(BlockContextIdPtrUniformSampledImage2dArray)

		ptr1D := b.EmitAccessChain(typePtrUniformSampledImage1D, idStaticTextures1D, bindingIndex)
		sampledImage1D := b.EmitLoad(typeSampledImage1D, ptr1D)
		ptr2D := b.EmitAccessChain(typePtrUniformSampledImage2D, idStaticTextures2D, bindingIndex)
		sampledImage2D := b.EmitLoad(typeSampledImage2D, ptr2D)
		ptr3D := b.EmitAccessChain(typePtrUniformSampledImage3D, idStaticTextures3D, bindingIndex)
		sampledImage3D := b.EmitLoad(typeSampledImage3D, ptr3D)
		ptr2DArray := b.EmitAccessChain(typePtrUniformSampledImage2DArray, idStaticTextures2DArray, bindingIndex)
		sampledImage2DArray := b.EmitLoad(typeSampledImage2DArray, ptr2DArray)

		// { body } - coordinates from VGPRs.
		coordX := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordZ := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+2))

		coord1D := coordX
		coord2D := b.EmitCompositeConstruct(typeV2Float, coordX, coordY)
		coord3D := b.EmitCompositeConstruct(typeV3Float, coordX, coordY, coordZ)
		coord2DArray := coord3D

		// Query LOD.
		resVec2 := EmitResourceTypeSwitch(b, typeV2Float, is1D, is3D, isCubeOrArray,
			sampledImage1D, sampledImage2D, sampledImage3D, sampledImage2DArray,
			coord1D, coord2D, coord3D, coord2DArray,
			func(imageId, coord SpirvId, resType int) SpirvId {
				queryCoord := coord
				if resType == ResType2DArray {
					coordU := b.EmitCompositeExtract(typeFloat, coord, 0)
					coordV := b.EmitCompositeExtract(typeFloat, coord, 1)
					queryCoord = b.EmitCompositeConstruct(typeV2Float, coordU, coordV)
				}

				// OpImageQueryLod returns a vec2 where X is clamped/accessed LOD, Y is raw LOD.
				return b.EmitImageQueryLod(typeV2Float, imageId, queryCoord)
			},
		)

		// Write results back to VGPRs based on dmask.
		// Dmask bit 0 = clamped/accessed LOD (resVec2.x).
		// Dmask bit 1 = raw/unclamped computed LOD (resVec2.y).
		vgprOffset := uint32(0)
		for i := range uint32(2) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec2, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	default:
		panic(fmt.Sprintf("unknown mimg op %s", gcnSpec.Mnemotics[gcnSpec.EncMIMG][details.Op]))
	}
}

const (
	ResType1D = iota
	ResType2D
	ResType3D
	ResType2DArray
)

// EmitResourceTypeSwitch generates structured control flow to switch based on the GCN resource type.
// The handler closure is called inside each specific basic block.
// If resultType is not 0, it merges the returned IDs using OpPhi and returns the final result.
func EmitResourceTypeSwitch(
	b *SpvBuilder,
	resultType SpirvId,
	is1D, is3D, isCubeOrArray SpirvId,
	image1D, image2D, image3D, image2DArray SpirvId,
	coord1D, coord2D, coord3D, coord2DArray SpirvId,
	handler func(imageId, coord SpirvId, resType int) SpirvId,
) SpirvId {
	mergeBlock1D := b.AllocId()
	mergeBlock3D := b.AllocId()
	mergeBlockArray := b.AllocId()

	block1D := b.AllocId()
	blockNot1D := b.AllocId()
	block3D := b.AllocId()
	blockNot3D := b.AllocId()
	block2DArray := b.AllocId()
	block2D := b.AllocId()

	var res1D, res2D, res3D, res2DArray SpirvId

	// Outer selection (is1D).
	b.EmitSelectionMerge(mergeBlock1D, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(is1D, block1D, blockNot1D)

	// 1D true.
	b.EmitLabel(block1D)
	res1D = handler(image1D, coord1D, ResType1D)
	b.EmitBranch(mergeBlock1D)

	// 1D false.
	b.EmitLabel(blockNot1D)
	b.EmitSelectionMerge(mergeBlock3D, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(is3D, block3D, blockNot3D)

	// 3D true.
	b.EmitLabel(block3D)
	res3D = handler(image3D, coord3D, ResType3D)
	b.EmitBranch(mergeBlock3D)

	// 3D false.
	b.EmitLabel(blockNot3D)
	b.EmitSelectionMerge(mergeBlockArray, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(isCubeOrArray, block2DArray, block2D)

	// Array true.
	b.EmitLabel(block2DArray)
	res2DArray = handler(image2DArray, coord2DArray, ResType2DArray)
	b.EmitBranch(mergeBlockArray)

	// Array false (2D).
	b.EmitLabel(block2D)
	res2D = handler(image2D, coord2D, ResType2D)
	b.EmitBranch(mergeBlockArray)

	// Merge blocks and Phis.
	b.EmitLabel(mergeBlockArray)
	var resArray SpirvId
	if resultType != 0 {
		resArray = b.EmitPhi(resultType, res2DArray, block2DArray, res2D, block2D)
	}
	b.EmitBranch(mergeBlock3D)

	b.EmitLabel(mergeBlock3D)
	var res3DMerge SpirvId
	if resultType != 0 {
		res3DMerge = b.EmitPhi(resultType, res3D, block3D, resArray, mergeBlockArray)
	}
	b.EmitBranch(mergeBlock1D)

	b.EmitLabel(mergeBlock1D)
	if resultType != 0 {
		return b.EmitPhi(resultType, res1D, block1D, res3DMerge, mergeBlock3D)
	}
	return 0
}

func extractResourceType(b *SpvBuilder, ctx *SpirvBlockContext, details *gcnSpec.MimgDetails) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	dw3 := ctx.GetGcnSgprId(b, details.Srsrc*4+3)
	shifted := b.EmitShiftRightLogical(typeUint, dw3, b.EmitConstantUint(typeUint, 28))
	return b.EmitBitwiseAnd(typeUint, shifted, b.EmitConstantUint(typeUint, 0xF))
}
