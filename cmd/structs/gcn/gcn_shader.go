package gcn

import "fmt"

type GcnShader struct {
	Instructions []Instruction
}

func NewGcnShader(dwords []uint32) (GcnShader, error) {
	shader := GcnShader{}
	i := 0
	for i < len(dwords) {
		enc, length := NewEncoding(dwords[i]), GetEncodingDwordLen(dwords[i])
		if i+length > len(dwords) {
			break
		}
		instr, err := NewInstruction(enc, dwords[i:i+length])
		if err != nil {
			return shader, err
		}

		shader.Instructions = append(shader.Instructions, instr)
		i += instr.DwordLen

		// S_ENDPGM (SOPP op=1) terminates the shader.
		if instr.Encoding == EncSOPP && instr.SOp == 1 {
			break
		}
	}

	return shader, nil
}

func (instr *Instruction) String() string {
	if instr.Encoding == EncUnknown {
		return fmt.Sprintf("%-6s  0x%08X                                   ; UNKNOWN", "?", instr.Dwords[0])
	}
	rawHex := fmt.Sprintf("0x%08X", instr.Dwords[0])
	if instr.DwordLen == 2 {
		rawHex += fmt.Sprintf(" 0x%08X", instr.Dwords[1])
	}

	return fmt.Sprintf("%-6s  %-22s  %-24s  %s", instr.Encoding, rawHex, instr.GetMnemotic(), instr.GetFieldsString())
}
