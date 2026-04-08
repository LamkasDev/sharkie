package gpu

import (
	"fmt"
	"os"
	"path"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/gookit/color"
)

// DumpShaderOnce scans the GCN shader at addr and prints its bytecode to the log.
func (l *Liverpool) DumpShaderOnce(address uintptr, stage string, rsrc1, rsrc2 uint32) error {
	if address == 0 {
		return fmt.Errorf("invalid address")
	}
	if _, loaded := l.SeenShaders.LoadOrStore(address, struct{}{}); loaded {
		return nil
	}

	// Scan guest memory for shader byte-code.
	dwords, foundEndpgm := scanShader(address)
	if dwords == nil {
		return fmt.Errorf("could not read memory")
	}
	if !foundEndpgm {
		logger.Printf("[%s] Hit cap on shader %s, skipping rest...",
			color.Blue.Sprint("SHADER"),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
	vgprs, sgprs := decodeRsrc1(rsrc1)
	scratchEnable := rsrc2 & 1
	userDataCount := (rsrc2 >> 1) & 0x1F
	logger.Printf("[%s] Scanned %s shader %s of %s bytes (vgprs=%s, sgprs=%s, scratchEnable=%s, userDataCount=%s, rsrc1=%s, rsrc2=%s)...\n",
		color.Blue.Sprint("SHADER"),
		color.Blue.Sprint(stage),
		color.Yellow.Sprintf("0x%X", address),
		color.Green.Sprint(len(dwords)*4),
		color.Yellow.Sprintf("0x%X", vgprs),
		color.Yellow.Sprintf("0x%X", sgprs),
		color.Yellow.Sprintf("0x%X", scratchEnable),
		color.Yellow.Sprintf("0x%X", userDataCount),
		color.Yellow.Sprintf("0x%X", rsrc1),
		color.Yellow.Sprintf("0x%X", rsrc2),
	)

	// Disassemble into GCN instructions.
	shader, err := gcn.NewGcnShader(dwords)
	if err != nil {
		panic(err)
	}
	var text string
	for i, instr := range shader.Instructions {
		instrText := instr.String()
		logger.Printf("[%s] %s: %s\n",
			color.Blue.Sprint("SHADER"),
			color.Green.Sprintf("%-4d", i),
			color.Cyan.Sprint(instrText),
		)
		text += fmt.Sprintf("%s\n", instrText)
	}

	// Dump the raw & disassembled shader.
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), len(dwords)*4)
	if err = os.MkdirAll(path.Join("temp", "shaders"), 0777); err != nil {
		return err
	}
	binFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.bin", address, stage))
	if err = os.WriteFile(binFilename, data, 0777); err != nil {
		return err
	}
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.txt", address, stage))
	if err = os.WriteFile(textFilename, []byte(text), 0777); err != nil {
		return err
	}
	logger.Printf("[%s] Dumped shader to %s...\n",
		color.Blue.Sprint("SHADER"),
		color.Blue.Sprint(binFilename),
	)

	return nil
}

// scanShader reads dwords from address until it finds S_ENDPGM or hits the limit.
func scanShader(address uintptr) (dwords []uint32, foundEndpgm bool) {
	dwords = make([]uint32, 0, 256)
	for i := 0; i < gcn.GcnShaderMaxDwords; i++ {
		dw := *(*uint32)(unsafe.Pointer(address + uintptr(i)*4))
		dwords = append(dwords, dw)
		if dw == gcn.GcnShaderEndProgram {
			return dwords, true
		}
	}

	return dwords, false
}

// decodeRsrc1 extracts VGPR and SGPR counts from a SPI_SHADER_PGM_RSRC1 or COMPUTE_PGM_RSRC1 value.
func decodeRsrc1(rsrc1 uint32) (vgprs, sgprs uint32) {
	vgprs = ((rsrc1 & 0x3F) + 1) * 4
	sgprs = (((rsrc1 >> 6) & 0xF) + 1) * 8

	return
}
