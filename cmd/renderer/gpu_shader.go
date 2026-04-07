package renderer

import (
	"fmt"
	"os"

	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
)

func loadShaderModule(dev vk.Device, path string) (vk.ShaderModule, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return vk.NullShaderModule, err
	}
	if len(code)%4 != 0 {
		return vk.NullShaderModule, fmt.Errorf("%s: size %d is not 4-byte aligned", path, len(code))
	}

	// Reinterpret []byte as []uint32 without copying.
	words := make([]uint32, len(code)/4)
	for i := range words {
		words[i] = uint32(code[i*4]) |
			uint32(code[i*4+1])<<8 |
			uint32(code[i*4+2])<<16 |
			uint32(code[i*4+3])<<24
	}

	var module vk.ShaderModule
	result := vk.CreateShaderModule(dev, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(code)),
		PCode:    words,
	}, nil, &module)
	return module, as.NewError(result)
}
