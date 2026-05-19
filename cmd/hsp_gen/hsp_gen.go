package main

import (
	"log"
	"os"
	"regexp"
	"strings"
)

// Post-processing step for hsp which adds uintptr support without weird workarounds.
func main() {
	for _, filePath := range os.Args[1:] {
		if !strings.Contains(filePath, "_gen.go") {
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		// Fix MarshalHash calls.
		reHash := regexp.MustCompile(`(?s)if oTemp, err := z\.([a-zA-Z0-9]+)\.MarshalHash\(\); err != nil \{\s+return nil, err\s+\} else \{\s+o = hsp\.AppendBytes\(o, oTemp\)\s+\}`)
		content = reHash.ReplaceAll(content, []byte("o = hsp.AppendUint64(o, uint64(z.$1))"))

		// Fix Msgsize calculations.
		reSize := regexp.MustCompile(`\+ \d+ \+ z\.([a-zA-Z0-9]+)\.Msgsize\(\)`)
		content = reSize.ReplaceAll(content, []byte("+ hsp.Uint64Size"))

		if err = os.WriteFile(filePath, []byte(content), 0644); err != nil {
			panic(err)
		}

		err = os.Remove(strings.Replace(filePath, "_gen.go", "_gen_test.go", 1))
		if err != nil {
			panic(err)
		}
	}
}
