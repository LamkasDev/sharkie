package gcn

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type GcnShaderStage uint8

const (
	GcnShaderStageVertex GcnShaderStage = iota
	GcnShaderStageHull
	GcnShaderStageEvaluation
	GcnShaderStageGeometry
	GcnShaderStageFragment
	GcnShaderStageCompute
)

var GcnShaderStages = []GcnShaderStage{
	GcnShaderStageVertex,
	GcnShaderStageHull,
	GcnShaderStageEvaluation,
	GcnShaderStageGeometry,
	GcnShaderStageFragment,
	GcnShaderStageCompute,
}

var GcnShaderStageNames = map[GcnShaderStage]string{
	GcnShaderStageVertex:     "VS",
	GcnShaderStageHull:       "HS",
	GcnShaderStageEvaluation: "ES",
	GcnShaderStageGeometry:   "GS",
	GcnShaderStageFragment:   "FS",
	GcnShaderStageCompute:    "CS",
}

func (stage GcnShaderStage) String() string {
	return GcnShaderStageNames[stage]
}

type GcnShader struct {
	Stage       GcnShaderStage
	Address     uintptr
	DwordLength uint64
	Cfg         GcnShaderCfg
}

func NewGcnShader(stage GcnShaderStage, address uintptr) (*GcnShader, error) {
	shader := &GcnShader{
		Stage:   stage,
		Address: address,
	}

	// Scan guest memory for shader byte-code.
	var dwords []uint32
	var foundEndProgram bool
	for i := uintptr(0); i < GcnShaderMaxDwords; i += 4 {
		dw := *(*uint32)(unsafe.Pointer(address + i))
		dwords = append(dwords, dw)
		if dw == GcnShaderEndProgram {
			foundEndProgram = true
			break
		}
	}
	if !foundEndProgram {
		logger.Printf("[%s] Hit cap on shader %s, skipping rest...",
			color.Blue.Sprint("SHADER"),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
	shader.DwordLength = uint64(len(dwords))

	// Disassemble the instructions.
	var instructions []spec.Instruction
	i := 0
	for i < len(dwords) {
		enc, length := NewEncoding(dwords[i]), GetEncodingDwordLen(dwords[i])
		if i+length > len(dwords) {
			break
		}
		instr, err := NewInstruction(uintptr(i), enc, dwords[i:i+length])
		if err != nil {
			return shader, err
		}

		instructions = append(instructions, instr)
		i += instr.DwordLen

		// S_ENDPGM (SOPP op=1) terminates the shader.
		if instr.Encoding == spec.EncSOPP && instr.Details.(*spec.ScalarDetails).Op == 1 {
			break
		}
	}

	// Build a control flow graph.
	var err error
	if shader.Cfg, err = NewGcnShaderCfg(instructions); err != nil {
		return shader, err
	}

	return shader, nil
}
