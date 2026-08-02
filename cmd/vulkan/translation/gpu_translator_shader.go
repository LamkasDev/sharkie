package translation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

type SpirvShaderKey struct {
	Address uintptr
	ThreadX uint32
	ThreadY uint32
	ThreadZ uint32

	PsInControl     uint32
	PsInputAddress  uint32
	PsInputControls [32]uint32

	FetchShaderAddress uintptr
	ClipDistEnable     uint8
	CullDistEnable     uint8

	DepthBeforeShader bool
	ZOrder            uint32
	BlendClamp        bool
	FrontFaceEnable   bool
}

func (t *GpuTranslator) GetShader(gcnShader *gcn.GcnShader) *spirv.SpirvShader {
	var ctx common.SpirvShaderContext
	switch gcnShader.Stage {
	case gcn.GcnShaderStageVertex:
		ctx = common.SpirvVertexShaderContext{}
	case gcn.GcnShaderStageFragment:
		ctx = common.SpirvFragmentShaderContext{}
	case gcn.GcnShaderStageCompute:
		ctx = common.SpirvComputeShaderContext{}
	}
	shader, _ := t.GetShaderWithContext(gcnShader, ctx)
	return shader
}

func (t *GpuTranslator) GetShaderAt(address uintptr) *spirv.SpirvShader {
	t.shadersMutex.Lock()
	defer t.shadersMutex.Unlock()
	for key, shader := range t.shaders {
		if key.Address == address {
			return shader
		}
	}

	return nil
}

func (t *GpuTranslator) GetShaderWithContext(gcnShader *gcn.GcnShader, context common.SpirvShaderContext) (*spirv.SpirvShader, SpirvShaderKey) {
	key := SpirvShaderKey{
		Address: gcnShader.Address,
	}
	switch c := context.(type) {
	case common.SpirvComputeShaderContext:
		key.ThreadX = c.ThreadX
		key.ThreadY = c.ThreadY
		key.ThreadZ = c.ThreadZ
	case common.SpirvFragmentShaderContext:
		key.PsInControl = c.PsInControl
		key.PsInputAddress = c.PsInputAddress
		key.PsInputControls = c.PsInputControls
		key.DepthBeforeShader = c.DepthBeforeShader
		key.ZOrder = c.ZOrder
		key.FrontFaceEnable = c.FrontFaceEnable
	case common.SpirvVertexShaderContext:
		key.FetchShaderAddress = c.FetchShaderAddress
		key.ClipDistEnable = c.ClipDistEnable
		key.CullDistEnable = c.CullDistEnable
	}

	// Get already loaded shader.
	t.shadersMutex.Lock()
	shader, ok := t.shaders[key]
	t.shadersMutex.Unlock()
	if ok {
		return shader, key
	}

	// Load the shader.
	t.shadersMutex.Lock()
	shader, err := spirv.NewSpirvShader(gcnShader, context)
	if err != nil {
		panic(err)
	}
	if err = t.DumpShaderOnce(key, shader); err != nil {
		panic(err)
	}
	t.shaders[key] = shader
	t.shadersMutex.Unlock()

	return shader, key
}

func (t *GpuTranslator) GetShaderModuleFromBytes(bytecode []uint32, name string) (vk.ShaderModule, error) {
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(bytecode) * 4),
		PCode:    bytecode,
	}, nil, &module)
	if err := vulkan.NewError(result); err != nil {
		return vk.NullShaderModule, err
	}
	vulkan.SetDebugUtilsObjectName(t.handles.Instance, t.handles.Device, vk.ObjectTypeShaderModule, uint64(uintptr(unsafe.Pointer(module))), name)

	return module, nil
}

func (t *GpuTranslator) GetShaderModule(key SpirvShaderKey, spirvShader *spirv.SpirvShader) (vk.ShaderModule, error) {
	// Get already created shader module.
	t.shaderModulesMutex.Lock()
	mod, ok := t.shaderModules[key]
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
	if err := vulkan.NewError(result); err != nil {
		return vk.NullShaderModule, fmt.Errorf("vkCreateShaderModule 0x%X: %w", spirvShader.GcnShader.Address, err)
	}
	vulkan.SetDebugUtilsObjectName(t.handles.Instance, t.handles.Device, vk.ObjectTypeShaderModule, uint64(uintptr(unsafe.Pointer(module))), fmt.Sprintf("Shader 0x%X", spirvShader.GcnShader.Address))
	t.shaderModulesMutex.Lock()
	t.shaderModules[key] = module
	t.shaderModulesMutex.Unlock()

	return module, nil
}

// DumpShaderOnce prints shader byte-code to a file.
func (t *GpuTranslator) DumpShaderOnce(key SpirvShaderKey, spirvShader *spirv.SpirvShader) error {
	// Check if tools available, otherwise skip.
	spirvValCheckCmd := exec.Command("spirv-val", "--help")
	if err := spirvValCheckCmd.Run(); err != nil {
		return nil
	}

	// Dump the recompiled shader.
	shaderDir := filepath.Join(config.GetGameCacheDir(), "shaders")
	if err := os.MkdirAll(shaderDir, 0755); err != nil {
		return err
	}
	textFilename := filepath.Join(shaderDir, fmt.Sprintf("shader_0x%X_%X_%s.spv", spirvShader.GcnShader.Address, key.FetchShaderAddress, spirvShader.GcnShader.Stage))
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
