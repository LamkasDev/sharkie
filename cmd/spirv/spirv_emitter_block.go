package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, cfg *GcnShaderCfg, block *GcnShaderCfgBlock, ctx *SpirvBlockContext) {
	// Start current block.
	b.EmitLabel(ctx.GetLabelId(block.Id))

	// Declare variables in entry block.
	if block.DwordOffset == 0 {
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		idC0 := ctx.GetConstId(ConstIdUint0)

		// Load user data buffer address from the push constant.
		b.EmitString("load user data buffer address")
		idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
		ptrBase := ctx.LoadPushConstantValue(b, PushConstantUserDataAddress)

		// Load 16 user data registers into s0-s15.
		b.EmitString("load user data registers")
		stageOffset := structs.GcnStageToUserDataOffset[ctx.Stage]
		for i := range uint32(16) {
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, ctx.GetConstId(SpirvId(stageOffset+i)))
			value := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ptr, spec.SpvMemoryAccessAligned, 4)
			ctx.SetGcnSgprId(b, i, value)
		}

		// Initialize EXEC and VCC.
		// EXEC is initialized to the subgroup's active mask.
		b.EmitString("initialize exec and vcc")
		typeV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
		idC3 := ctx.GetConstId(ConstIdUint3) // Subgroup
		ballot := b.EmitGroupNonUniformBallot(typeV4Uint, idC3, ctx.GetId(BlockContextIdTrue))
		execLo := b.EmitCompositeExtract(typeUint, ballot, 0)
		execHi := b.EmitCompositeExtract(typeUint, ballot, 1)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, execLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, execHi)

		// VCC is initialized to 0.
		b.EmitString("set vcc to 0")
		ctx.StoreRegisterPointer(b, gcnSpec.OpVccLo, idC0)
		ctx.StoreRegisterPointer(b, gcnSpec.OpVccHi, idC0)

		// Load vertex index and instance index into v0 and v1.
		if ctx.Stage == GcnShaderStageVertex {
			b.EmitString("load vertex and instance index")
			v0 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))
			ctx.SetGcnVgprId(b, 0, v0)
			v1 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdInstanceIndex))
			ctx.SetGcnVgprId(b, 1, v1)

			// Inline fetch shader instructions.
			if len(ctx.Context.(SpirvVertexShaderContext).FetchShaderInstructions) > 0 {
				b.EmitString("inline fetch shader loads")
				for _, instr := range ctx.Context.(SpirvVertexShaderContext).FetchShaderInstructions {
					emitInstruction(b, instr, ctx)
				}
			}
		}

		// Initialize input VGPRs for fragment shader according to SPI_PS_INPUT_ADDR.
		if ctx.Stage == GcnShaderStageFragment {
			b.EmitString("initialize fragment input vgprs")
			typeUint := ctx.GetId(BlockContextIdTypeUint)
			typeFloat := ctx.GetId(BlockContextIdTypeFloat)
			idHalfF := ctx.GetConstId(ConstIdFloat05)
			idOneF := ctx.GetConstId(ConstIdFloat1)
			idZeroU := ctx.GetConstId(ConstIdUint0)
			idHalfU := b.EmitBitcast(typeUint, idHalfF)
			idOneU := b.EmitBitcast(typeUint, idOneF)

			ctxFs, ok := ctx.Context.(SpirvFragmentShaderContext)
			dstVgpr := uint32(0)

			if ok {
				psAddr := ctxFs.PsInputAddress
				// 0. persp_sample_ena (2 VGPRs)
				if psAddr&(1<<0) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 1. persp_center_ena (2 VGPRs)
				if psAddr&(1<<1) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 2. persp_centroid_ena (2 VGPRs)
				if psAddr&(1<<2) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 3. persp_pull_model_ena (3 VGPRs: I/W, J/W, 1/W)
				if psAddr&(1<<3) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idOneU)
					dstVgpr++
				}
				// 4. linear_sample_ena (2 VGPRs)
				if psAddr&(1<<4) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 5. linear_center_ena (2 VGPRs)
				if psAddr&(1<<5) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 6. linear_centroid_ena (2 VGPRs)
				if psAddr&(1<<6) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
					ctx.SetGcnVgprId(b, dstVgpr, idHalfU)
					dstVgpr++
				}
				// 7. line_stipple_tex_ena (1 VGPR)
				if psAddr&(1<<7) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idZeroU)
					dstVgpr++
				}
				// 8-11. pos_x/y/z/w_float_ena
				hasFragCoord := (psAddr>>8)&0xF != 0
				var fragX, fragY, fragZ, fragW SpirvId
				if hasFragCoord {
					fragCoordPtr := ctx.GetId(BlockContextIdFragCoord)
					if fragCoordPtr != 0 {
						fragCoordVal := b.EmitLoad(ctx.GetId(BlockContextIdTypeV4Float), fragCoordPtr)
						fragX = b.EmitBitcast(typeUint, b.EmitCompositeExtract(typeFloat, fragCoordVal, 0))
						fragY = b.EmitBitcast(typeUint, b.EmitCompositeExtract(typeFloat, fragCoordVal, 1))
						fragZ = b.EmitBitcast(typeUint, b.EmitCompositeExtract(typeFloat, fragCoordVal, 2))
						rawW := b.EmitCompositeExtract(typeFloat, fragCoordVal, 3)
						recipW := b.EmitFDiv(typeFloat, idOneF, rawW)
						fragW = b.EmitBitcast(typeUint, recipW)
					}
				}
				if psAddr&(1<<8) != 0 {
					val := idZeroU
					if fragX != 0 {
						val = fragX
					}
					ctx.SetGcnVgprId(b, dstVgpr, val)
					dstVgpr++
				}
				if psAddr&(1<<9) != 0 {
					val := idZeroU
					if fragY != 0 {
						val = fragY
					}
					ctx.SetGcnVgprId(b, dstVgpr, val)
					dstVgpr++
				}
				if psAddr&(1<<10) != 0 {
					val := idZeroU
					if fragZ != 0 {
						val = fragZ
					}
					ctx.SetGcnVgprId(b, dstVgpr, val)
					dstVgpr++
				}
				if psAddr&(1<<11) != 0 {
					val := idZeroU
					if fragW != 0 {
						val = fragW
					}
					ctx.SetGcnVgprId(b, dstVgpr, val)
					dstVgpr++
				}
				// 12. front_face_ena (1 VGPR)
				if psAddr&(1<<12) != 0 {
					frontFacingPtr := ctx.GetId(BlockContextIdFrontFacing)
					val := idZeroU
					if frontFacingPtr != 0 {
						frontFacingBool := b.EmitLoad(ctx.GetId(BlockContextIdTypeBool), frontFacingPtr)
						val = b.EmitSelect(typeUint, frontFacingBool, ctx.GetConstId(ConstIdUint1), idZeroU)
					}
					ctx.SetGcnVgprId(b, dstVgpr, val)
					dstVgpr++
				}
				// 13. ancillary_ena (1 VGPR)
				if psAddr&(1<<13) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idZeroU)
					dstVgpr++
				}
				// 14. sample_coverage_ena (1 VGPR)
				if psAddr&(1<<14) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idZeroU)
					dstVgpr++
				}
				// 15. pos_fixed_pt_ena (1 VGPR)
				if psAddr&(1<<15) != 0 {
					ctx.SetGcnVgprId(b, dstVgpr, idZeroU)
					dstVgpr++
				}
			}

			// Fallback: if no barycentrics or system inputs were allocated, provide default barycentrics
			if dstVgpr == 0 {
				ctx.SetGcnVgprId(b, 0, idHalfU)
				ctx.SetGcnVgprId(b, 1, idHalfU)
				ctx.SetGcnVgprId(b, 2, idOneU)
			}

			b.EmitStore(ctx.GetId(BlockContextIdIsValidPixel), ctx.GetId(BlockContextIdTrue))
		}

		// Initialize thread IDs for compute shader.
		if ctx.Stage == GcnShaderStageCompute {
			typePtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
			b.EmitString("initialize compute thread ids")

			// Built-in workgroup ID.
			workgroupVec := b.EmitLoad(ctx.GetId(BlockContextIdTypeV3Uint), ctx.GetId(BlockContextIdWorkgroupId))

			// Loaded user data registers.
			userSgprCount := ctx.LoadPushConstantValue(b, PushConstantUserSgprCount)
			// rsrc2 := ctx.LoadPushConstantValue(b, PushConstantShaderRsrc2)
			sgprIdx := userSgprCount

			// TGID.X (bit 7)
			// condX := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, rsrc2, ctx.GetConstId(ConstIdUint7)), ctx.GetConstId(ConstIdUint1))
			ptrX := b.EmitAccessChain(typePtrFnUint, ctx.GcnSgprArrayId, sgprIdx)
			wgXRaw := b.EmitCompositeExtract(typeUint, workgroupVec, 0)
			wgXFinal := wgXRaw
			if ctx.Context.(SpirvComputeShaderContext).ThreadX > 1024 {
				splitFactor := b.EmitConstantUint(typeUint, (ctx.Context.(SpirvComputeShaderContext).ThreadX+1023)/1024)
				wgXFinal = b.EmitUDiv(typeUint, wgXRaw, splitFactor)
			}
			b.EmitStore(ptrX, wgXFinal)
			sgprIdx = b.EmitIAdd(typeUint, sgprIdx, ctx.GetConstId(ConstIdUint1))

			// TGID.Y (bit 8)
			// condY := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, rsrc2, ctx.GetConstId(ConstIdUint8)), ctx.GetConstId(ConstIdUint1))
			ptrY := b.EmitAccessChain(typePtrFnUint, ctx.GcnSgprArrayId, sgprIdx)
			b.EmitStore(ptrY, b.EmitCompositeExtract(typeUint, workgroupVec, 1))
			sgprIdx = b.EmitIAdd(typeUint, sgprIdx, ctx.GetConstId(ConstIdUint1))

			// TGID.Z (bit 9)
			// condZ := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, rsrc2, ctx.GetConstId(ConstIdUint9)), ctx.GetConstId(ConstIdUint1))
			ptrZ := b.EmitAccessChain(typePtrFnUint, ctx.GcnSgprArrayId, sgprIdx)
			b.EmitStore(ptrZ, b.EmitCompositeExtract(typeUint, workgroupVec, 2))
			sgprIdx = b.EmitIAdd(typeUint, sgprIdx, ctx.GetConstId(ConstIdUint1))

			// TG_SIZE_EN (bit 10)
			// condTgSize := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, rsrc2, ctx.GetConstId(ConstIdUint10)), ctx.GetConstId(ConstIdUint1))
			ptrTgSize := b.EmitAccessChain(typePtrFnUint, ctx.GcnSgprArrayId, sgprIdx)
			b.EmitStore(ptrTgSize, ctx.GetConstId(ConstIdUint0))
			sgprIdx = b.EmitIAdd(typeUint, sgprIdx, ctx.GetConstId(ConstIdUint1))

			// Builtin local invocation ID.
			localVec := b.EmitLoad(ctx.GetId(BlockContextIdTypeV3Uint), ctx.GetId(BlockContextIdLocalInvocationId))
			localXRaw := b.EmitCompositeExtract(typeUint, localVec, 0)
			localXFinal := localXRaw
			if ctx.Context.(SpirvComputeShaderContext).ThreadX > 1024 {
				// Restore original Local ID.
				splitFactor := b.EmitConstantUint(typeUint, (ctx.Context.(SpirvComputeShaderContext).ThreadX+1023)/1024)
				wgMod := b.EmitUMod(typeUint, wgXRaw, splitFactor)
				offset := b.EmitIMul(typeUint, wgMod, b.EmitConstantUint(typeUint, 1024))
				localXFinal = b.EmitIAdd(typeUint, localXRaw, offset)
			}

			ctx.SetGcnVgprId(b, 0, localXFinal)
			ctx.SetGcnVgprId(b, 1, b.EmitCompositeExtract(typeUint, localVec, 1))
			ctx.SetGcnVgprId(b, 2, b.EmitCompositeExtract(typeUint, localVec, 2))
		}
	}

	// Reset condition ID.
	ctx.GcnConditionId = ctx.GetId(BlockContextIdFalse)

	// We split the header by immediately merging and branching to a body block.
	if block.IsLoopHeader {
		bodyLabel := b.AllocId()
		mergeLabel := ctx.GetLabelId(block.MergeBlockId)

		var continueLabel SpirvId
		if block.ContinueBlockId == block.Id {
			// Self-loop (needs a dummy continue block).
			continueLabel = ctx.GetId(BlockContextIdContinueBlocks + SpirvId(block.Id))
		} else {
			// Normal loop (use the real continue block).
			continueLabel = ctx.GetLabelId(block.ContinueBlockId)
		}

		b.EmitLoopMerge(mergeLabel, continueLabel, spec.SpvLoopControlNone)
		b.EmitBranch(bodyLabel)
		b.EmitLabel(bodyLabel)
	}

	// Emit instructions for current block.
	for i := range block.Instructions {
		emitInstruction(b, &block.Instructions[i], ctx)
	}

	// Terminate current block.
	switch block.Term {
	case TermCBranch:
		EmitConditionalBranch(b, cfg, block, ctx)
	case TermBranch, TermFallthrough:
		if len(block.Successors) > 0 {
			targetId := block.Successors[0]
			targetSpvId := ctx.GetLabelId(targetId)

			// Intercept self-loop back-edges.
			if targetId == block.Id && block.IsLoopHeader && block.ContinueBlockId == block.Id {
				targetSpvId = ctx.GetId(BlockContextIdContinueBlocks + SpirvId(targetId))
			}

			b.EmitBranch(targetSpvId)
		} else {
			b.EmitUnreachable()
		}
	case TermEndpgm, TermExpDone:
		switch ctx.Stage {
		case GcnShaderStageVertex:
			if true {
				break
			}
			formatId := b.EmitString("Vertex %d: pos=(%f, %f, %f, %f) param_out=(%f, %f, %f, %f)\n")
			typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
			typeFloat := ctx.GetId(BlockContextIdTypeFloat)
			posId := b.EmitLoad(typeV4Float, ctx.GetId(BlockContextIdPosOut))
			paramOut := b.EmitLoad(typeV4Float, ctx.GetId(BlockContextIdParamOut0))
			vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

			px := b.EmitCompositeExtract(typeFloat, posId, 0)
			py := b.EmitCompositeExtract(typeFloat, posId, 1)
			pz := b.EmitCompositeExtract(typeFloat, posId, 2)
			pw := b.EmitCompositeExtract(typeFloat, posId, 3)

			pox := b.EmitCompositeExtract(typeFloat, paramOut, 0)
			poy := b.EmitCompositeExtract(typeFloat, paramOut, 1)
			poz := b.EmitCompositeExtract(typeFloat, paramOut, 2)
			pow := b.EmitCompositeExtract(typeFloat, paramOut, 3)

			b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdTypeDebugPrintf), 1,
				formatId, vertexIndexId, px, py, pz, pw, pox, poy, poz, pow)
		case GcnShaderStageFragment:
			if true {
				break
			}
			formatId := b.EmitString(fmt.Sprintf("Fragment 0x%X: color=(%%f, %%f, %%f, %%f)\n", ctx.Address))
			typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
			typeFloat := ctx.GetId(BlockContextIdTypeFloat)
			colorId := b.EmitLoad(typeV4Float, ctx.GetId(BlockContextIdColorOut0))

			cx := b.EmitCompositeExtract(typeFloat, colorId, 0)
			cy := b.EmitCompositeExtract(typeFloat, colorId, 1)
			cz := b.EmitCompositeExtract(typeFloat, colorId, 2)
			cw := b.EmitCompositeExtract(typeFloat, colorId, 3)

			b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdTypeDebugPrintf), 1,
				formatId, cx, cy, cz, cw)
		}
		// ctx.EmitDebugPrintRegisters(b)
		if ctx.Stage == GcnShaderStageFragment {
			isValidPixel := b.EmitLoad(ctx.GetId(BlockContextIdTypeBool), ctx.GetId(BlockContextIdIsValidPixel))
			isInvalid := b.EmitLogicalNot(ctx.GetId(BlockContextIdTypeBool), isValidPixel)

			killLabel := b.AllocId()
			mergeLabel := b.AllocId()

			b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
			b.EmitBranchConditional(isInvalid, killLabel, mergeLabel)

			b.EmitLabel(killLabel)
			b.EmitKill()

			b.EmitLabel(mergeLabel)
		}
		b.EmitReturn()
	default:
		b.EmitReturn()
	}

	// Emit the explicit continue block for self-loops.
	if block.IsLoopHeader && block.ContinueBlockId == block.Id {
		continueSpvId := ctx.GetId(BlockContextIdContinueBlocks + SpirvId(block.Id))
		b.EmitLabel(continueSpvId)
		b.EmitBranch(ctx.GetLabelId(block.Id)) // Branch back to the true header.
	}
}
