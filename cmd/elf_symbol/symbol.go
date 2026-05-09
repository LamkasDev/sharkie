package symbol

import (
	"bufio"
	"os"
	"strings"
)

// https://github.com/OpenOrbis/OpenOrbis-PS4-Toolchain/wiki/PS4-ELF-Specification---Dynlib-Data#nid-table
const nidEncoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-"

var symbolMap = make(map[string]string)

// LoadSymbolMap loads the symbol map from the given CSV file (ex. aerolib.csv).
func LoadSymbolMap(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			mangled := strings.TrimSpace(parts[0])
			readable := strings.TrimSpace(parts[1])
			symbolMap[mangled] = readable
		}
	}
}

// MangledToReadable returns the readable name for a mangled symbol.
// If not found, returns the mangled name.
func MangledToReadable(mangled string) string {
	// Strip suffix starting with # (e.g. kxXCvcat1cM#r#q -> kxXCvcat1cM)
	baseName := mangled
	if idx := strings.Index(mangled, "#"); idx != -1 {
		baseName = mangled[:idx]
	}

	if readable, ok := symbolMap[baseName]; ok {
		return readable
	}
	return mangled
}

// DecodeNidChar returns index of encoded NID character.
func DecodeNidChar(c byte) uint16 {
	return uint16(strings.IndexByte(nidEncoding, c))
}
