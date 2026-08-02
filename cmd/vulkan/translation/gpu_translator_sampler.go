package translation

import (
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vkGcn "github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetSampler(descriptor spirvStructs.SamplerDescriptor) (vk.Sampler, error) {
	hash := descriptor.Hash()

	// Get already created sampler.
	t.samplersMutex.Lock()
	defer t.samplersMutex.Unlock()
	if sampler, ok := t.samplers[hash]; ok {
		return sampler, nil
	}

	// Prepare parameters for sampler.
	anisotropyEnable := vk.Bool32(vk.False)
	if descriptor.MaxAnisoRatio > 0 {
		anisotropyEnable = vk.True
	}

	// Create the sampler.
	var sampler vk.Sampler
	result := vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:            vk.StructureTypeSamplerCreateInfo,
		MagFilter:        vkGcn.TranslateFilter(descriptor.XyMagFilter),
		MinFilter:        vkGcn.TranslateFilter(descriptor.XyMinFilter),
		MipmapMode:       vkGcn.TranslateMipmapMode(descriptor.ZFilter),
		AddressModeU:     vkGcn.TranslateClampMode(descriptor.ClampX),
		AddressModeV:     vkGcn.TranslateClampMode(descriptor.ClampY),
		AddressModeW:     vkGcn.TranslateClampMode(descriptor.ClampZ),
		AnisotropyEnable: anisotropyEnable,
		MaxAnisotropy:    float32(descriptor.MaxAnisoRatio),
		MaxLod:           descriptor.MaxLod,
		BorderColor:      vkGcn.TranslateBorderColorType(descriptor.BorderColorType),
	}, nil, &sampler)
	if err := vulkan.NewError(result); err != nil {
		return vk.NullSampler, err
	}
	t.samplers[hash] = sampler

	return sampler, nil
}
