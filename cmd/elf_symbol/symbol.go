package symbol

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/LamkasDev/sharkie"
)

// https://github.com/OpenOrbis/OpenOrbis-PS4-Toolchain/wiki/PS4-ELF-Specification---Dynlib-Data#nid-table
const nidEncoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-"

var GoNicEncoding = base64.NewEncoding(nidEncoding).WithPadding(base64.NoPadding)

var symbolMap = make(map[string]string)

// LoadSymbolMap loads the symbol map from the given CSV file (ex. aerolib.csv).
func LoadSymbolMap(path string) {
	file, err := sharkie.Assets.Open(path)
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

// ExtractNidBase safely extracts the 11-character NID base from a symbol name.
func ExtractNidBase(mangled string) string {
	// A valid NID with suffixes (hash#lib#mod or hash_lib_mod) is at least 15 characters.
	// We check the 11th and 13th indices (0-based) to see if they are matching separators.
	if len(mangled) >= 15 {
		sep1 := mangled[11]
		sep2 := mangled[13]
		if sep1 == '#' && sep2 == '#' {
			return mangled[:11]
		}
	}

	// Fallback for strings that might only have a single # suffix.
	if idx := strings.Index(mangled, "#"); idx != -1 {
		return mangled[:idx]
	}

	// Return as-is (might be an 11-char NID with no suffix or a regular string).
	return mangled
}

// MangledToReadable returns the readable name for a mangled symbol (hash).
// If not found, returns the mangled name.
func MangledToReadable(mangled string) string {
	baseName := ExtractNidBase(mangled)
	if readable, ok := symbolMap[baseName]; ok {
		return readable
	}

	return mangled
}

// DecodeNidChar returns index of encoded NID character.
func DecodeNidChar(c byte) uint16 {
	return uint16(strings.IndexByte(nidEncoding, c))
}

// ReadableToMangled returns the mangled symbol (hash) for a readable name.
func ReadableToMangled(symbol string) string {
	salt, _ := hex.DecodeString("518D64A635DED8C1E6B039B1C3E55230")
	data := append([]byte(symbol), salt...)
	hash := sha1.Sum(data)
	nidBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		nidBytes[i] = hash[7-i]
	}
	encoded := GoNicEncoding.EncodeToString(nidBytes)

	return encoded
}
