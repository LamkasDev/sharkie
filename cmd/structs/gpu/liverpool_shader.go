package gpu

import (
	"fmt"
	"os"
	"path"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// DumpShaderOnce scans the GCN shader at addr and prints its bytecode to the log.
func (l *Liverpool) DumpShaderOnce(address uintptr, stage string, rsrc1, rsrc2 uint32) error {
	if address == 0 {
		return fmt.Errorf("invalid address")
	}
	if _, loaded := l.SeenShaders.LoadOrStore(address, struct{}{}); loaded {
		return nil
	}

	dwords, _ := scanShader(address)
	if dwords == nil {
		return fmt.Errorf("could not read memory")
	}
	vgprs, sgprs := decodeRsrc1(rsrc1)
	shaderSize := len(dwords) * 4
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), shaderSize)

	filename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.bin", address, stage))
	if err := os.MkdirAll(path.Join("temp", "shaders"), 0777); err != nil {
		return err
	}
	err := os.WriteFile(filename, data, 0777)
	if err != nil {
		return err
	}

	logger.Printf("[SHADER] Dumped %s | VGPRs: %d, SGPRs: %d, Size: %d bytes\n", filename, vgprs, sgprs, shaderSize)
	return nil
}

// scanShader reads dwords from addr until it finds S_ENDPGM or hits the limit.
func scanShader(address uintptr) (dwords []uint32, endpgmIdx int) {
	if address == 0 {
		return nil, 0
	}

	dwords = make([]uint32, 0, 256)
	for i := 0; i < gcn.GcnShaderMaxDwords; i++ {
		dw := *(*uint32)(unsafe.Pointer(address + uintptr(i)*4))
		dwords = append(dwords, dw)
		if dw == gcn.GcnShaderEndProgram {
			return dwords, i
		}
	}

	return dwords, -1
}

// decodeRsrc1 extracts VGPR and SGPR counts from a SPI_SHADER_PGM_RSRC1 or COMPUTE_PGM_RSRC1 value.
func decodeRsrc1(rsrc1 uint32) (vgprs, sgprs uint32) {
	vgprs = ((rsrc1 & 0x3F) + 1) * 4
	sgprs = (((rsrc1 >> 6) & 0xF) + 1) * 8

	return
}
