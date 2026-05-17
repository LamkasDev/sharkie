package structs

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
	"github.com/gookit/color"
	"go101.org/nstd"
)

type BufferDescriptor struct {
	BaseAddress   uintptr // 48 bits
	Stride        uint16  // 14 bits
	SwizzleCache  bool    // 1 bit
	SwizzleEnable bool    // 1 bit
	Records       uint32  // 32 bits
	DstSelX       uint8   // 3 bits
	DstSelY       uint8   // 3 bits
	DstSelZ       uint8   // 3 bits
	DstSelW       uint8   // 3 bits
	NumFormat     uint8   // 3 bits
	DataFormat    uint8   // 4 bits
	ElementSize   uint8   // 2 bits (decoded as 2, 4, 8, 16)
	IndexStride   uint8   // 2 bits (decoded as 8, 16, 32, 64)
	AddTidEnable  bool
	Atc           bool
	HashEnable    bool
	Heap          bool
	MType         uint8 // 3 bits
	Type          uint8 // 2 bits
}

func NewBufferDescriptor(dw0, dw1, dw2, dw3 uint32) BufferDescriptor {
	d := BufferDescriptor{}

	// DW0 & DW1: Base Address (48 bits) + Stride (14) + Flags
	d.BaseAddress = uintptr(dw0) | (uintptr(dw1&0xFFFF) << 32)
	d.Stride = uint16((dw1 >> 16) & 0x3FFF)
	d.SwizzleCache = (dw1 >> 30 & 1) == 1
	d.SwizzleEnable = (dw1 >> 31 & 1) == 1

	// DW2: Num_records (32 bits)
	d.Records = dw2

	// DW3: Formatting and Flags
	d.DstSelX = uint8(dw3 & 0x7)
	d.DstSelY = uint8((dw3 >> 3) & 0x7)
	d.DstSelZ = uint8((dw3 >> 6) & 0x7)
	d.DstSelW = uint8((dw3 >> 9) & 0x7)
	d.NumFormat = uint8((dw3 >> 12) & 0x7)
	d.DataFormat = uint8((dw3 >> 15) & 0xF)

	// Element Size mapping: 0=2, 1=4, 2=8, 3=16
	d.ElementSize = 2 << ((dw3 >> 19) & 0x3)
	// Index Stride mapping: 0=8, 1=16, 2=32, 3=64
	d.IndexStride = 8 << ((dw3 >> 21) & 0x3)

	d.AddTidEnable = (dw3 >> 23 & 1) == 1
	d.Atc = (dw3 >> 24 & 1) == 1
	d.HashEnable = (dw3 >> 25 & 1) == 1
	d.Heap = (dw3 >> 26 & 1) == 1
	d.MType = uint8((dw3 >> 27) & 0x7)
	d.Type = uint8((dw3 >> 30) & 0x3)

	return d
}

func (d BufferDescriptor) Print() {
	logger.Printf("base=%s, stride=%s, records=%s, swizzle=%s, dst=%s, formats=%s, addressing=%s, flags=%s, memory=%s\n",
		color.Yellow.Sprintf("0x%X", d.BaseAddress),
		color.Green.Sprint(d.Stride),
		color.Green.Sprint(d.Records),
		color.Green.Sprintf("(en=%d cache=%d)", nstd.Btoi(d.SwizzleEnable), nstd.Btoi(d.SwizzleCache)),
		color.Green.Sprintf("(%d %d %d %d)", d.DstSelX, d.DstSelY, d.DstSelZ, d.DstSelW),
		color.Green.Sprintf("(num=%d data=%d)", d.NumFormat, d.DataFormat),
		color.Green.Sprintf("(elemSize=%d indexStride=%d)", d.ElementSize, d.IndexStride),
		color.Green.Sprintf("(addTid=%d atc=%d hash=%d heap=%d)", nstd.Btoi(d.AddTidEnable), nstd.Btoi(d.Atc), nstd.Btoi(d.HashEnable), nstd.Btoi(d.Heap)),
		color.Green.Sprintf("(mtype=%d type=%d)", d.MType, d.Type),
	)
}

type BufferResource struct {
	BaseAddress  SpirvId
	Stride       SpirvId
	NumRecords   SpirvId
	SwizzleEn    SpirvId
	ElementSize  SpirvId
	IndexStride  SpirvId
	AddTidEnable SpirvId
	Dw3          SpirvId
}

func NewBufferResource(b *SpvBuilder, ctx *SpirvBlockContext, srsrc uint32) BufferResource {
	sgprBase := srsrc * 4
	dw0 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase)
	dw1 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+1)
	dw2 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+2)
	dw3 := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+sgprBase+3)

	return BufferResource{
		BaseAddress:  GetResourceBaseAddress(b, ctx, dw0, dw1),
		Stride:       GetResourceStride(b, ctx, dw1),
		NumRecords:   GetResourceNumRecords(b, ctx, dw2),
		SwizzleEn:    GetResourceSwizzleEnable(b, ctx, dw1),
		ElementSize:  GetResourceElementSize(b, ctx, dw3),
		IndexStride:  GetResourceIndexStride(b, ctx, dw3),
		AddTidEnable: GetResourceAddTidEnable(b, ctx, dw3),
		Dw3:          dw3,
	}
}

// GetResourceBaseAddress extracts the base address from T# dword 0 and 1.
func GetResourceBaseAddress(b *SpvBuilder, ctx *SpirvBlockContext, dw0, dw1 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	baseLo := dw0
	baseHi := b.EmitBitwiseAnd(typeUint, dw1, ctx.GetConstId(ConstIdUintFFFF))
	base := ctx.Pack64(b, baseLo, baseHi)
	return base
}

func GetResourceStride(b *SpvBuilder, ctx *SpirvBlockContext, dw1 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	shifted := b.EmitShiftRightLogical(typeUint, dw1, ctx.GetConstId(ConstIdUint16))
	// Mask to 14 bits [13:0] to strip CACHE_SWIZZLE (bit 14) and SWIZZLE_EN (bit 15).
	return b.EmitBitwiseAnd(typeUint, shifted, ctx.GetConstId(ConstIdUint3FFF))
}

func GetResourceNumRecords(b *SpvBuilder, ctx *SpirvBlockContext, dw2 SpirvId) SpirvId {
	return dw2
}

func GetResourceSwizzleEnable(b *SpvBuilder, ctx *SpirvBlockContext, dw1 SpirvId) SpirvId {
	return ctx.TestMask(b, dw1, 1<<31)
}

func GetResourceAddTidEnable(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	return ctx.TestMask(b, dw3, 1<<23)
}

func GetResourceElementSize(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	bits := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint19), ctx.GetConstId(ConstIdUint2))
	// 0=2, 1=4, 2=8, 3=16. This is 2 << bits.
	return b.EmitShiftLeftLogical(typeUint, ctx.GetConstId(ConstIdUint2), bits)
}

func GetResourceIndexStride(b *SpvBuilder, ctx *SpirvBlockContext, dw3 SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	bits := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint21), ctx.GetConstId(ConstIdUint2))
	// 0=8, 1=16, 2=32, 3=64. This is 8 << bits.
	return b.EmitShiftLeftLogical(typeUint, ctx.GetConstId(ConstIdUint8), bits)
}

// CalculateBufferOffset calculates the byte offset into a buffer resource according to linear or swizzled addressing.
func CalculateBufferOffset(b *SpvBuilder, ctx *SpirvBlockContext, stride, swizzleEn, elementSize, indexStride, addTidEnable, index, offset SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	// index = index + (addTidEnable ? thread_id[5:0] : 0)
	threadId := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
	threadId = b.EmitBitwiseAnd(typeUint, threadId, ctx.GetConstId(ConstIdUint63))
	index = b.EmitSelect(typeUint, addTidEnable, b.EmitIAdd(typeUint, index, threadId), index)

	// Linear: buffer_offset = index * stride + offset
	linearOffset := b.EmitIAdd(typeUint, b.EmitIMul(typeUint, index, stride), offset)

	// Swizzled:
	// index_msb = index / indexStride
	// index_lsb = index % indexStride
	// offset_msb = offset / elementSize
	// offset_lsb = offset % elementSize
	// buffer_offset = (index_msb * stride + offset_msb * elementSize) * indexStride + index_lsb * elementSize + offset_lsb
	indexMsb := b.EmitUDiv(typeUint, index, indexStride)
	indexLsb := b.EmitUMod(typeUint, index, indexStride)
	offsetMsb := b.EmitUDiv(typeUint, offset, elementSize)
	offsetLsb := b.EmitUMod(typeUint, offset, elementSize)

	term1 := b.EmitIAdd(typeUint, b.EmitIMul(typeUint, indexMsb, stride), b.EmitIMul(typeUint, offsetMsb, elementSize))
	swizzledOffset := b.EmitIAdd(typeUint,
		b.EmitIAdd(typeUint, b.EmitIMul(typeUint, term1, indexStride), b.EmitIMul(typeUint, indexLsb, elementSize)),
		offsetLsb)

	return b.EmitSelect(typeUint, swizzleEn, swizzledOffset, linearOffset)
}

// CalculateBufferRangeCheck returns a boolean ID which is true if the access is out of range.
func CalculateBufferRangeCheck(b *SpvBuilder, ctx *SpirvBlockContext, res BufferResource, sgprOffset, index, offset, bufferOffset, idxenOrAddTidEnable SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)

	strideIsZero := b.EmitIEqual(typeBool, res.Stride, ctx.GetConstId(ConstIdUint0))

	// Case 1: stride == 0
	// outOfRange = bufferOffset + sgprOffset >= num_records
	totalOffsetStrideZero := b.EmitIAdd(typeUint, bufferOffset, sgprOffset)
	outOfRangeStrideZero := b.EmitUGreaterThanEqual(typeBool, totalOffsetStrideZero, res.NumRecords)

	// Case 2: stride != 0
	// outOfRange = index >= num_records || ((idxen | addTidEnable) && offset >= stride)
	indexInRange := b.EmitULessThan(typeBool, index, res.NumRecords)
	indexOutOfRange := b.EmitLogicalNot(typeBool, indexInRange)

	offsetInRange := b.EmitULessThan(typeBool, offset, res.Stride)
	offsetOutOfRange := b.EmitLogicalNot(typeBool, offsetInRange)

	cond2 := b.EmitLogicalAnd(typeBool, idxenOrAddTidEnable, offsetOutOfRange)
	outOfRangeStrideNotZero := b.EmitLogicalOr(typeBool, indexOutOfRange, cond2)

	return b.EmitSelect(typeBool, strideIsZero, outOfRangeStrideZero, outOfRangeStrideNotZero)
}
