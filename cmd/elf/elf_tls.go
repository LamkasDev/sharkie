package elf

import "encoding/binary"

type ElfTlsSection struct {
	ImageVirtualAddress uint64
	InitImageSize       uint64
	ImageSize           uint64
	Align               uint64

	Offset      uint64
	ModuleIndex uint64
}

// NewTlsSection loads the PT_TLS section at offset.
func (e *Elf) NewTlsSection(data []byte, offset uint64) *ElfTlsSection {
	pVaddr := binary.LittleEndian.Uint64(data[offset+0x10:])
	pFilesz := binary.LittleEndian.Uint64(data[offset+0x20:])
	pMemsz := binary.LittleEndian.Uint64(data[offset+0x28:])
	pAlign := binary.LittleEndian.Uint64(data[offset+0x30:])
	imageSize := GetAlignedSize(pMemsz, pAlign)

	return &ElfTlsSection{
		ImageVirtualAddress: pVaddr,
		InitImageSize:       pFilesz,
		ImageSize:           imageSize,
		Align:               pAlign,
	}
}
