package elf

import "encoding/binary"

// ElfTlsSection represents the Thread Local Storage (TLS) program header.
type ElfTlsSection struct {
	ImageVirtualAddress uint64 // Virtual address of the TLS initialization image.
	InitImageSize       uint64 // Size of the TLS initialization image in the file.
	ImageSize           uint64 // Total size of the TLS segment in memory (including uninitialized data).
	Align               uint64 // Alignment of the TLS segment.

	Offset uint64 // Offset of the TLS segment within the ELF file.
}

// NewTlsSection creates a new ElfTlsSection by parsing the PT_TLS section of an ELF file.
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
