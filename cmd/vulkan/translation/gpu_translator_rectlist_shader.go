package translation

import (
	"os"

	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	vk "github.com/goki/vulkan"
)

//go:generate glslc --target-env=vulkan1.2 ../../../data/shaders/rectlist.tesc -o ../../../data/shaders/rectlist_tesc.spv
//go:generate glslc --target-env=vulkan1.2 ../../../data/shaders/rectlist.tese -o ../../../data/shaders/rectlist_tese.spv

func (t *GpuTranslator) getAuxShaderModule(path, name string, cache *vk.ShaderModule) (vk.ShaderModule, error) {
	t.shaderModulesMutex.Lock()
	defer t.shaderModulesMutex.Unlock()
	if *cache != vk.NullShaderModule {
		return *cache, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return vk.NullShaderModule, err
	}
	module, err := t.GetShaderModuleFromBytes(common.SpvBytesToWords(bytes), name)
	if err != nil {
		return vk.NullShaderModule, err
	}
	*cache = module

	return module, nil
}

func (t *GpuTranslator) GetRectlistTescShader() (vk.ShaderModule, error) {
	return t.getAuxShaderModule("data/shaders/rectlist_tesc.spv", "Rectlist TECS", &t.rectlistTescShaderModule)
}

func (t *GpuTranslator) GetRectlistTeseShader() (vk.ShaderModule, error) {
	return t.getAuxShaderModule("data/shaders/rectlist_tese.spv", "Rectlist TESE", &t.rectlistTeseShaderModule)
}
