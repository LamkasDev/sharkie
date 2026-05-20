package vulkan

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	vk "github.com/goki/vulkan"
)

type SpirvShaderKey struct {
	Address uintptr
	ThreadX uint32
	ThreadY uint32
	ThreadZ uint32

	PsInputAddress  uint32
	PsInputControls [32]uint32
}

func (t *GpuTranslator) GetShader(gcnShader *gcn.GcnShader) *spirv.SpirvShader {
	return t.GetShaderWithContext(gcnShader, spirv.SpirvShaderContext{})
}

func (t *GpuTranslator) GetShaderWithContext(gcnShader *gcn.GcnShader, context spirv.SpirvShaderContext) *spirv.SpirvShader {
	key := SpirvShaderKey{
		Address: gcnShader.Address,
		ThreadX: context.ThreadX,
		ThreadY: context.ThreadY,
		ThreadZ: context.ThreadZ,

		PsInputAddress:  context.PsInputAddress,
		PsInputControls: context.PsInputControls,
	}

	// Get already loaded shader.
	t.shadersMutex.Lock()
	shader, ok := t.shaders[key]
	t.shadersMutex.Unlock()
	if ok {
		return shader
	}

	// Load the shader.
	t.shadersMutex.Lock()
	shader, err := spirv.NewSpirvShader(gcnShader, context)
	if err != nil {
		panic(err)
	}
	if err = t.DumpShaderOnce(shader); err != nil {
		panic(err)
	}
	t.shaders[key] = shader
	t.shadersMutex.Unlock()

	return shader
}

func (t *GpuTranslator) GetShaderModuleFromBytes(bytecode []uint32) (vk.ShaderModule, error) {
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(bytecode) * 4),
		PCode:    bytecode,
	}, nil, &module)
	if err := as.NewError(result); err != nil {
		return vk.NullShaderModule, err
	}

	return module, nil
}

func (t *GpuTranslator) GetShaderModule(spirvShader *spirv.SpirvShader) (vk.ShaderModule, error) {
	// Get already created shader module.
	t.shaderModulesMutex.Lock()
	mod, ok := t.shaderModules[spirvShader.Address]
	t.shaderModulesMutex.Unlock()
	if ok {
		return mod, nil
	}

	// Create the shader module.
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(spirvShader.Code) * 4),
		PCode:    spirvShader.Code,
	}, nil, &module)
	if err := as.NewError(result); err != nil {
		return vk.NullShaderModule, fmt.Errorf("vkCreateShaderModule 0x%X: %w", spirvShader.Address, err)
	}
	t.shaderModulesMutex.Lock()
	t.shaderModules[spirvShader.Address] = module
	t.shaderModulesMutex.Unlock()

	return module, nil
}

func (t *GpuTranslator) GetRectlistShader() (vk.ShaderModule, error) {
	t.shaderModulesMutex.Lock()
	defer t.shaderModulesMutex.Unlock()
	if t.rectlistShaderModule != vk.NullShaderModule {
		return t.rectlistShaderModule, nil
	}

	bytes, err := os.ReadFile("temp/shaders/shader_rectlist.spv")
	if err != nil {
		return vk.NullShaderModule, err
	}
	module, err := t.GetShaderModuleFromBytes(common.SpvBytesToWords(bytes))
	if err != nil {
		return vk.NullShaderModule, err
	}
	t.rectlistShaderModule = module

	return module, nil
}

// DumpShaderOnce prints shader byte-code to a file.
func (t *GpuTranslator) DumpShaderOnce(spirvShader *spirv.SpirvShader) error {
	// Check if tools available, otherwise skip.
	spirvValCheckCmd := exec.Command("spirv-val", "--help")
	if err := spirvValCheckCmd.Run(); err != nil {
		return nil
	}

	// Dump the recompiled shader.
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.spv", spirvShader.Address, spirvShader.Stage))
	if err := os.WriteFile(textFilename, common.SpvWordsToBytes(spirvShader.Code), 0777); err != nil {
		return err
	}

	// Validate it.
	spirvValCmd := exec.Command("spirv-val", textFilename)
	spirvValCmd.Stderr = os.Stderr
	if err := spirvValCmd.Run(); err != nil {
		return err
	}

	// TODO: run spirv-opt.

	return nil
}
