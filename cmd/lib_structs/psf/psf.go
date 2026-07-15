package psf

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"unsafe"
)

const (
	PsfMagic      = uint32(0x00505346)
	PsfVersion1_1 = uint32(0x00000101)
	PsfVersion1_0 = uint32(0x00000100)
)

type PsfEntryFmt uint16

const (
	PsfEntryFmtBinary  = PsfEntryFmt(0x0004) // Binary data.
	PsfEntryFmtText    = PsfEntryFmt(0x0204) // String in UTF-8 format and NULL terminated.
	PsfEntryFmtInteger = PsfEntryFmt(0x0404) // Signed 32-bit integer.
)

type Psf struct {
	Entries     []PsfEntry
	MapBinaries map[string][]byte
	MapStrings  map[string]string
	MapIntegers map[string]int32
}

type PsfEntry struct {
	Key      string
	ParamFmt PsfEntryFmt
	MaxLen   uint32
}

// Original structs.
type PsfHeader struct {
	Magic             uint32 // u32_be
	Version           uint32 // u32_le
	KeyTableOffset    uint32 // u32_le
	DataTableOffset   uint32 // u32_le
	IndexTableEntries uint32 // u32_le
}

const PsfHeaderSize = unsafe.Sizeof(PsfHeader{})

type PsfRawEntry struct {
	KeyOffset   uint16 // u16_le
	ParamFmt    uint16 // u16_be
	ParamLen    uint32 // u32_le
	ParamMaxLen uint32 // u32_le
	Dataffset   uint32 // u32_le
}

const PsfRawEntrySize = unsafe.Sizeof(PsfRawEntry{})

func NewPsf() *Psf {
	return &Psf{
		Entries:     []PsfEntry{},
		MapBinaries: map[string][]byte{},
		MapStrings:  map[string]string{},
		MapIntegers: map[string]int32{},
	}
}

func NewPsfFromPath(filepath string) (*Psf, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("file at %s is empty", filepath)
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	psf, err := NewPsfFromData(data)
	if err != nil {
		return nil, err
	}

	return psf, nil
}

func NewPsfFromData(data []byte) (*Psf, error) {
	if len(data) < int(PsfHeaderSize) {
		return nil, fmt.Errorf("buffer too small for psf header")
	}
	magic := binary.BigEndian.Uint32(data[0:4])
	if magic != PsfMagic {
		return nil, fmt.Errorf("invalid psf magic number: 0x%08x", magic)
	}
	version := binary.LittleEndian.Uint32(data[4:8])
	if version != PsfVersion1_1 && version != PsfVersion1_0 {
		return nil, fmt.Errorf("unsupported psf version: 0x%08x", version)
	}

	// Read header.
	keyTableOffset := binary.LittleEndian.Uint32(data[8:12])
	dataTableOffset := binary.LittleEndian.Uint32(data[12:16])
	numEntries := binary.LittleEndian.Uint32(data[16:20])

	rawEntriesOffset := uint32(PsfHeaderSize)
	if uint32(len(data)) < rawEntriesOffset+(numEntries*16) {
		return nil, fmt.Errorf("buffer too small for psf entry table")
	}

	// Iterate raw entries.
	psf := NewPsf()
	for i := uint32(0); i < numEntries; i++ {
		entryOffset := rawEntriesOffset + (i * 16)

		// Read raw entry.
		keyOffset := binary.LittleEndian.Uint16(data[entryOffset : entryOffset+2])
		paramFmt := PsfEntryFmt(binary.LittleEndian.Uint16(data[entryOffset+2 : entryOffset+4]))
		paramLen := binary.LittleEndian.Uint32(data[entryOffset+4 : entryOffset+8])
		paramMaxLen := binary.LittleEndian.Uint32(data[entryOffset+8 : entryOffset+12])
		dataOffset := binary.LittleEndian.Uint32(data[entryOffset+12 : entryOffset+16])

		// Read null-terminated string from the key table.
		keyStart := keyTableOffset + uint32(keyOffset)
		if keyStart >= uint32(len(data)) {
			return nil, fmt.Errorf("key offset out of bounds")
		}
		keyEnd := keyStart
		for keyEnd < uint32(len(data)) && data[keyEnd] != 0 {
			keyEnd++
		}
		key := string(data[keyStart:keyEnd])

		// Append entry.
		entry := PsfEntry{
			Key:      key,
			ParamFmt: paramFmt,
			MaxLen:   paramMaxLen,
		}
		psf.Entries = append(psf.Entries, entry)

		// Slice out the raw data for this parameter.
		valStart := dataTableOffset + dataOffset
		valEnd := valStart + paramLen
		if valEnd > uint32(len(data)) {
			return nil, fmt.Errorf("psf entry '%s' data out of bounds", key)
		}
		valData := data[valStart:valEnd]

		switch paramFmt {
		case PsfEntryFmtBinary:
			binVal := make([]byte, paramLen)
			copy(binVal, valData)
			psf.MapBinaries[key] = binVal
		case PsfEntryFmtText:
			// Text format is UTF-8 and NULL terminated, so we trim up to the NULL byte.
			strLen := 0
			for strLen < len(valData) && valData[strLen] != 0 {
				strLen++
			}
			psf.MapStrings[key] = string(valData[:strLen])
		case PsfEntryFmtInteger:
			if paramLen != 4 { // sizeof(s32)
				return nil, fmt.Errorf("psf integer entry '%s' size mismatch (expected=4, got=%d)", key, paramLen)
			}
			psf.MapIntegers[key] = int32(binary.LittleEndian.Uint32(valData))
		default:
			return nil, fmt.Errorf("unknown psf entry format: 0x%04x for key '%s'", paramFmt, key)
		}
	}

	return psf, nil
}

func (psf *Psf) Encode() ([]byte, error) {
	// We need to sort entries alphabetically to comply with spec.
	sort.Slice(psf.Entries, func(i, j int) bool {
		return psf.Entries[i].Key < psf.Entries[j].Key
	})

	// Prepare buffer.
	numEntries := len(psf.Entries)
	data := make([]byte, int(PsfHeaderSize)+(int(PsfRawEntrySize)*numEntries))

	// Write header.
	binary.BigEndian.PutUint32(data[0:4], PsfMagic)                // u32_be magic
	binary.LittleEndian.PutUint32(data[4:8], PsfVersion1_1)        // u32_le version
	binary.LittleEndian.PutUint32(data[16:20], uint32(numEntries)) // u32_le index_table_entries

	// Write key table.
	keyTableOffset := len(data)
	binary.LittleEndian.PutUint32(data[8:12], uint32(keyTableOffset)) // u32_le key_table_offset in header
	for i, entry := range psf.Entries {
		rawEntryOffset := int(PsfHeaderSize) + (i * int(PsfRawEntrySize))
		binary.LittleEndian.PutUint16(data[rawEntryOffset:rawEntryOffset+2], uint16(len(data)-keyTableOffset)) // u16_le key_offset
		binary.LittleEndian.PutUint16(data[rawEntryOffset+2:rawEntryOffset+4], uint16(entry.ParamFmt))         // u16_le param_fmt
		binary.LittleEndian.PutUint32(data[rawEntryOffset+8:rawEntryOffset+12], entry.MaxLen)                  // u32_le param_max_len
		data = append(data, []byte(entry.Key)...)
		data = append(data, 0) // NULL terminator.
	}

	// 4-byte alignment padding for data table.
	if len(data)%4 != 0 {
		padding := 4 - (len(data) % 4)
		data = append(data, make([]byte, padding)...)
	}

	// Write data table.
	dataTableOffset := len(data)
	binary.LittleEndian.PutUint32(data[12:16], uint32(dataTableOffset)) // u32_le data_table_offset in header
	for i, entry := range psf.Entries {
		// 4-byte alignment padding for data.
		if len(data)%4 != 0 {
			padding := 4 - (len(data) % 4)
			data = append(data, make([]byte, padding)...)
		}

		// Write entry.
		rawEntryOffset := int(PsfHeaderSize) + (i * int(PsfRawEntrySize))
		binary.LittleEndian.PutUint32(data[rawEntryOffset+12:rawEntryOffset+16], uint32(len(data)-dataTableOffset)) // data_offset (u32_le)

		// Append actual data.
		var paramLen uint32
		switch entry.ParamFmt {
		case PsfEntryFmtBinary:
			val, ok := psf.MapBinaries[entry.Key]
			if !ok {
				return nil, fmt.Errorf("missing binary data for entry index %d", i)
			}
			paramLen = uint32(len(val))
			data = append(data, val...)
		case PsfEntryFmtText:
			val, ok := psf.MapStrings[entry.Key]
			if !ok {
				return nil, fmt.Errorf("missing string data for entry index %d", i)
			}
			paramLen = uint32(len(val) + 1)
			data = append(data, []byte(val)...)
			data = append(data, 0) // NULL terminator.
		case PsfEntryFmtInteger:
			val, ok := psf.MapIntegers[entry.Key]
			if !ok {
				return nil, fmt.Errorf("missing integer data for entry index %d", i)
			}
			paramLen = 4 // sizeof(s32)
			intBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(intBytes, uint32(val))
			data = append(data, intBytes...)
		default:
			return nil, fmt.Errorf("unknown psf entry format: %v", entry.ParamFmt)
		}
		binary.LittleEndian.PutUint32(data[rawEntryOffset+4:rawEntryOffset+8], paramLen) // u32_le param_len in raw entry

		// Verify max length constraint and pad.
		if int(entry.MaxLen) < int(paramLen) {
			return nil, fmt.Errorf("psf entry max size mismatch: key '%s' len %d exceeds max %d", entry.Key, paramLen, entry.MaxLen)
		}
		additionalPadding := entry.MaxLen - paramLen
		if additionalPadding > 0 {
			data = append(data, make([]byte, additionalPadding)...)
		}
	}

	return data, nil
}
