package renderer

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
)

// DumpShaderOnce prints shader byte-code to a file.
func (t *GpuTranslator) DumpShaderOnce(shader *SpirvShader) error {
	// Dump the recompiled shader.
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.spv", shader.Address, shader.Stage))
	if err := os.WriteFile(textFilename, SpvWordsToBytes(shader.Code), 0777); err != nil {
		return err
	}
	cmd := exec.Command("spirv-dis", textFilename)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
