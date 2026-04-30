package renderer

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetShader(drawShader *gcn.GcnShader) *spirv.SpirvShader {
	// Get already loaded shader.
	t.shadersMutex.Lock()
	shader, ok := t.shaders[drawShader.Address]
	t.shadersMutex.Unlock()
	if ok {
		return shader
	}

	// Load the shader.
	t.shadersMutex.Lock()
	shader, err := spirv.NewSpirvShader(drawShader, spirv.SpirvShaderContext{})
	if err != nil {
		panic(err)
	}
	if err = t.DumpShaderOnce(shader); err != nil {
		panic(err)
	}
	t.shaders[drawShader.Address] = shader
	t.shadersMutex.Unlock()

	return shader
}

func (t *GpuTranslator) GetShaderModule(shader *spirv.SpirvShader) (vk.ShaderModule, error) {
	// Get already created shader module.
	t.shaderModulesMutex.Lock()
	mod, ok := t.shaderModules[shader.Address]
	t.shaderModulesMutex.Unlock()
	if ok {
		return mod, nil
	}

	// Create the shader module.
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(shader.Code) * 4),
		PCode:    shader.Code,
	}, nil, &module)
	if err := as.NewError(result); err != nil {
		return vk.NullShaderModule, fmt.Errorf("vkCreateShaderModule 0x%X: %w", shader.Address, err)
	}
	t.shaderModulesMutex.Lock()
	t.shaderModules[shader.Address] = module
	t.shaderModulesMutex.Unlock()

	return module, nil
}

// DumpShaderOnce prints shader byte-code to a file.
func (t *GpuTranslator) DumpShaderOnce(shader *spirv.SpirvShader) error {
	// Dump the recompiled shader.
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.spv", shader.Address, shader.Stage))
	if err := os.WriteFile(textFilename, spirv.SpvWordsToBytes(shader.Code), 0777); err != nil {
		return err
	}
	cmd := exec.Command("spirv-dis", textFilename)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
