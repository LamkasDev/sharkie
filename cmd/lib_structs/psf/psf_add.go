package psf

import (
	"encoding/binary"
	"fmt"
)

var maxSizes = map[string]uint32{
	"ACCOUNT_ID":         8,
	"CATEGORY":           4,
	"DETAIL":             1024,
	"FORMAT":             4,
	"MAINTITLE":          128,
	"PARAMS":             1024,
	"SAVEDATA_BLOCKS":    8,
	"SAVEDATA_DIRECTORY": 32,
	"SUBTITLE":           128,
	"TITLE_ID":           12,
}

func getMaxKeySize(key string, defaultValue uint32) uint32 {
	if size, exists := maxSizes[key]; exists {
		return size
	}
	return defaultValue
}

func (psf *Psf) AddBinary(key string, value []byte, update bool) error {
	index, found := psf.findEntry(key)
	if found && !update {
		return fmt.Errorf("tried to add binary key that already exists: %s", key)
	}

	maxLen := getMaxKeySize(key, uint32(len(value)))
	if found {
		if psf.Entries[index].ParamFmt != PsfEntryFmtBinary {
			return fmt.Errorf("format change is not supported")
		}
		psf.Entries[index].MaxLen = maxLen
		psf.MapBinaries[index] = value
		return nil
	}

	psf.Entries = append(psf.Entries, PsfEntry{
		Key:      key,
		ParamFmt: PsfEntryFmtBinary,
		MaxLen:   maxLen,
	})
	psf.MapBinaries[len(psf.Entries)-1] = value
	return nil
}

func (psf *Psf) AddBinaryUint64(key string, value uint64, update bool) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	return psf.AddBinary(key, data, update)
}

func (psf *Psf) AddString(key string, value string, update bool) error {
	index, found := psf.findEntry(key)
	if found && !update {
		return fmt.Errorf("tried to add string key that already exists: %s", key)
	}

	maxLen := getMaxKeySize(key, uint32(len(value)+1)) // +1 for NULL terminator.
	if found {
		if psf.Entries[index].ParamFmt != PsfEntryFmtText {
			return fmt.Errorf("format change is not supported")
		}
		psf.Entries[index].MaxLen = maxLen
		psf.MapStrings[index] = value
		return nil
	}

	psf.Entries = append(psf.Entries, PsfEntry{
		Key:      key,
		ParamFmt: PsfEntryFmtText,
		MaxLen:   maxLen,
	})
	psf.MapStrings[len(psf.Entries)-1] = value
	return nil
}

// AddInteger adds or updates an integer value.
func (psf *Psf) AddInteger(key string, value int32, update bool) error {
	index, found := psf.findEntry(key)
	if found && !update {
		return fmt.Errorf("tried to add integer key that already exists: %s", key)
	}

	maxLen := uint32(4) // sizeof(s32).
	if found {
		if psf.Entries[index].ParamFmt != PsfEntryFmtInteger {
			return fmt.Errorf("format change is not supported")
		}
		psf.Entries[index].MaxLen = maxLen
		psf.MapIntegers[index] = value
		return nil
	}

	psf.Entries = append(psf.Entries, PsfEntry{
		Key:      key,
		ParamFmt: PsfEntryFmtInteger,
		MaxLen:   maxLen,
	})
	psf.MapIntegers[len(psf.Entries)-1] = value
	return nil
}
