package psf

import "fmt"

func (psf *Psf) findEntry(key string) (int, bool) {
	for i, entry := range psf.Entries {
		if entry.Key == key {
			return i, true
		}
	}

	return -1, false
}

func (psf *Psf) GetBinary(key string) ([]byte, bool) {
	index, found := psf.findEntry(key)
	if !found {
		return nil, false
	}
	if psf.Entries[index].ParamFmt != PsfEntryFmtBinary {
		fmt.Printf("mismatched psf entry format (expected=%d, got=%d)", PsfEntryFmtBinary, psf.Entries[index].ParamFmt)
		return nil, false
	}
	return psf.MapBinaries[index], true
}

func (psf *Psf) GetString(key string) (string, bool) {
	index, found := psf.findEntry(key)
	if !found {
		return "", false
	}
	if psf.Entries[index].ParamFmt != PsfEntryFmtText {
		fmt.Printf("mismatched psf entry format (expected=%d, got=%d)", PsfEntryFmtText, psf.Entries[index].ParamFmt)
		return "", false
	}
	return psf.MapStrings[index], true
}

func (psf *Psf) GetInteger(key string) (int32, bool) {
	index, found := psf.findEntry(key)
	if !found {
		return 0, false
	}
	if psf.Entries[index].ParamFmt != PsfEntryFmtInteger {
		fmt.Printf("mismatched psf entry format (expected=%d, got=%d)", PsfEntryFmtInteger, psf.Entries[index].ParamFmt)
		return 0, false
	}
	return psf.MapIntegers[index], true
}
