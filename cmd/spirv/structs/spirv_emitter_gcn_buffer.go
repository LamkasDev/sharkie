package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/gookit/color"
	"go101.org/nstd"
)

func NewBufferDescriptor(dwords []uint32) BufferDescriptor {
	baseAddress := (uintptr(dwords[0]) | (uintptr(dwords[1]&0xFFFF) << 32)) & 0xFFFFFFFFFF
	return BufferDescriptor{
		BaseAddress:   baseAddress,
		Stride:        uint16((dwords[1] >> 16) & 0x3FFF),
		SwizzleCache:  (dwords[1]>>30)&1 == 1,
		SwizzleEnable: (dwords[1]>>31)&1 == 1,

		NumRecords:   dwords[2],
		DstSelX:      uint8(dwords[3] & 0x7),
		DstSelY:      uint8((dwords[3] >> 3) & 0x7),
		DstSelZ:      uint8((dwords[3] >> 6) & 0x7),
		DstSelW:      uint8((dwords[3] >> 9) & 0x7),
		NumFormat:    uint8((dwords[3] >> 12) & 0x7),
		DataFormat:   uint8((dwords[3] >> 15) & 0xF),
		ElementSize:  2 << ((dwords[3] >> 19) & 0x3),
		IndexStride:  8 << ((dwords[3] >> 21) & 0x3),
		AddTidEnable: (dwords[3]>>23)&1 == 1,
		Atc:          (dwords[3]>>24)&1 == 1,
		HashEnable:   (dwords[3]>>25)&1 == 1,
		Heap:         (dwords[3]>>26)&1 == 1,
		MType:        uint8((dwords[3] >> 27) & 0x7),
		Type:         uint8((dwords[3] >> 30) & 0x3),
	}
}

func (d BufferDescriptor) Print() {
	logger.Printf("base=%s, stride=%s, records=%s, swizzle=%s, dst=%s, formats=%s, addressing=%s, flags=%s, memory=%s\n",
		color.Yellow.Sprintf("0x%X", d.BaseAddress),
		color.Green.Sprint(d.Stride),
		color.Green.Sprint(d.NumRecords),
		color.Green.Sprintf("(en=%d cache=%d)", nstd.Btoi(d.SwizzleEnable), nstd.Btoi(d.SwizzleCache)),
		color.Green.Sprintf("(%d %d %d %d)", d.DstSelX, d.DstSelY, d.DstSelZ, d.DstSelW),
		color.Green.Sprintf("(num=%d data=%d)", d.NumFormat, d.DataFormat),
		color.Green.Sprintf("(elemSize=%d indexStride=%d)", d.ElementSize, d.IndexStride),
		color.Green.Sprintf("(addTid=%d atc=%d hash=%d heap=%d)", nstd.Btoi(d.AddTidEnable), nstd.Btoi(d.Atc), nstd.Btoi(d.HashEnable), nstd.Btoi(d.Heap)),
		color.Green.Sprintf("(mtype=%d type=%d)", d.MType, d.Type),
	)
}

type BufferResource struct {
	Stride        SpirvId
	SwizzleEnable SpirvId
	NumRecords    SpirvId

	ElementSize  SpirvId
	IndexStride  SpirvId
	AddTidEnable SpirvId
	Dw3          SpirvId
}

func NewBufferResource(b *SpvBuilder, ctx *SpirvBlockContext, srsrc uint32) BufferResource {
	sgprBase := srsrc * 4
	dw1 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+1)
	dw2 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+2)
	dw3 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+3)

	return BufferResource{
		Stride:        GetResourceStride(b, ctx, dw1),
		SwizzleEnable: GetResourceSwizzleEnable(b, ctx, dw1),
		NumRecords:    GetResourceNumRecords(b, ctx, dw2),

		ElementSize:  GetResourceElementSize(b, ctx, dw3),
		IndexStride:  GetResourceIndexStride(b, ctx, dw3),
		AddTidEnable: GetResourceAddTidEnable(b, ctx, dw3),
		Dw3:          dw3,
	}
}

func GetResourceStride(b *SpvBuilder, ctx *SpirvBlockContext, dw1 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	shifted := b.EmitShiftRightLogical(typeUint, dw1, ctx.GetConstId(ConstIdUint16))
	return b.EmitBitwiseAnd(typeUint, shifted, ctx.GetConstId(ConstIdUint3FFF))
}

func GetResourceSwizzleEnable(b *SpvBuilder, ctx *SpirvBlockContext, dw1 SpirvId) SpirvId {
	return ctx.TestMask(b, dw1, 1<<31)
}

func GetResourceNumRecords(b *SpvBuilder, ctx *SpirvBlockContext, dw2 SpirvId) SpirvId {
	return dw2
}

func GetResourceElementSize(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	bits := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint19), ctx.GetConstId(ConstIdUint2))
	return b.EmitShiftLeftLogical(typeUint, ctx.GetConstId(ConstIdUint2), bits)
}

func GetResourceIndexStride(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	bits := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint21), ctx.GetConstId(ConstIdUint2))
	return b.EmitShiftLeftLogical(typeUint, ctx.GetConstId(ConstIdUint8), bits)
}

func GetResourceAddTidEnable(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	return ctx.TestMask(b, dw3, 1<<23)
}

// CalculateBufferOffset calculates the byte offset into a buffer resource according to linear or swizzled addressing.
func CalculateBufferOffset(b *SpvBuilder, ctx *SpirvBlockContext, res BufferResource, index, offset SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	// index = index + (addTidEnable ? thread_id[5:0] : 0)
	threadId := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
	threadId = b.EmitBitwiseAnd(typeUint, threadId, ctx.GetConstId(ConstIdUint63))
	index = b.EmitSelect(typeUint, res.AddTidEnable, b.EmitIAdd(typeUint, index, threadId), index)

	// 8.1.5.1 Linear Buffer Addressing
	// buffer_offset = index * stride + offset
	linearOffset := b.EmitIAdd(typeUint, b.EmitIMul(typeUint, index, res.Stride), offset)

	// 8.1.5.2 Swizzled Buffer Addressing
	// index_msb = index / indexStride
	// index_lsb = index % indexStride
	// offset_msb = offset / elementSize
	// offset_lsb = offset % elementSize
	// buffer_offset = (index_msb * stride + offset_msb * elementSize) * indexStride + index_lsb * elementSize + offset_lsb
	indexMsb := b.EmitUDiv(typeUint, index, res.IndexStride)
	indexLsb := b.EmitUMod(typeUint, index, res.IndexStride)
	offsetMsb := b.EmitUDiv(typeUint, offset, res.ElementSize)
	offsetLsb := b.EmitUMod(typeUint, offset, res.ElementSize)

	term1 := b.EmitIAdd(typeUint, b.EmitIMul(typeUint, indexMsb, res.Stride), b.EmitIMul(typeUint, offsetMsb, res.ElementSize))
	swizzledOffset := b.EmitIAdd(typeUint,
		b.EmitIAdd(typeUint,
			b.EmitIMul(typeUint, term1, res.IndexStride),
			b.EmitIMul(typeUint, indexLsb, res.ElementSize),
		),
		offsetLsb)

	return b.EmitSelect(typeUint, res.SwizzleEnable, swizzledOffset, linearOffset)
}

// CalculateBufferRangeCheck returns a boolean ID which is true if the access is out of range.
func CalculateBufferRangeCheck(b *SpvBuilder, ctx *SpirvBlockContext, res BufferResource, index, offset, bufferOffset, sgprOffset, idxenOrAddTidEnable SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)

	strideIsZero := b.EmitIEqual(typeBool, res.Stride, ctx.GetConstId(ConstIdUint0))

	// If ((const_stride == 0) && (buffer_offset >= (const_num_records – sgpr_offset))
	// – If op is a write or atomic, drop the write.
	// – If op is a read or atomic, return 0.
	totalOffsetStrideZero := b.EmitIAdd(typeUint, bufferOffset, sgprOffset)
	outOfRangeStrideZero := b.EmitUGreaterThanEqual(typeBool, totalOffsetStrideZero, res.NumRecords)

	// If (const_stride != 0 && ((index >= const_num_records) || ((inst_idxen | const_add_tid_enable) && (offset >= const_stride)))
	// – If op is a write or atomic, drop the write.
	// – If op is a read or atomic, return 0.
	indexOutOfRange := b.EmitUGreaterThanEqual(typeBool, index, res.NumRecords)
	offsetOutOfRange := b.EmitUGreaterThanEqual(typeBool, offset, res.Stride)
	offsetCheckWithFlags := b.EmitLogicalAnd(typeBool, idxenOrAddTidEnable, offsetOutOfRange)
	outOfRangeStrideNotZero := b.EmitLogicalOr(typeBool, indexOutOfRange, offsetCheckWithFlags)

	return b.EmitSelect(typeBool, strideIsZero, outOfRangeStrideZero, outOfRangeStrideNotZero)
}
