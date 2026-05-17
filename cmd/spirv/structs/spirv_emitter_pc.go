package structs

import "unsafe"

type PushConstants struct {
	UserDataAddress         uint64
	OnionMemoryBaseAddress  uint64
	GarlicMemoryBaseAddress uint64

	TexelBuffer0FormatSize uint32
	TexelBuffer1FormatSize uint32
	TexelBuffer2FormatSize uint32
	TexelBuffer3FormatSize uint32

	TexelBuffer0FormatStride uint32
	TexelBuffer1FormatStride uint32
	TexelBuffer2FormatStride uint32
	TexelBuffer3FormatStride uint32

	UserSgprCount uint32
	ShaderRsrc2   uint32
}

const PushConstantsSize = uint32(unsafe.Sizeof(PushConstants{}))
