package renderer

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) loadStubShaders() error {
	var err error
	var vertModule vk.ShaderModule
	vertModule, err = loadShaderModule(t.handles.Device, "data/shaders/stub_vert.spv")
	if err != nil {
		return fmt.Errorf("stub_vert.spv: %w", err)
	}
	t.stubVertShader = vertModule
	var fragModule vk.ShaderModule
	fragModule, err = loadShaderModule(t.handles.Device, "data/shaders/stub_frag.spv")
	if err != nil {
		return fmt.Errorf("stub_frag.spv: %w", err)
	}
	t.stubFragShader = fragModule

	return nil
}

// DumpShaderOnce prints shader byte-code to a file.
func (t *GpuTranslator) DumpShaderOnce(shader *SpirvShader) error {
	// Dump the recompiled shader.
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.spv", shader.Address, shader.Stage))
	if err := os.WriteFile(textFilename, SpvWordsToBytes(shader.Code), 0777); err != nil {
		return err
	}
	asmFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.spvasm", shader.Address, shader.Stage))
	cmd := exec.Command("spirv-dis", textFilename, "--no-indent", "--offsets", "--comment", "-o", asmFilename)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
