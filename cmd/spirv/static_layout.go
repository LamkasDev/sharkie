package spirv

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func BuildStaticLayout(resources []SpirvShaderResource, shader *gcn.GcnShader) []ShaderResourceBinding {
	var layout []ShaderResourceBinding
	var sampledIndex, storageIndex uint32

	// Offset bindings for vertex shaders so they don't overlap with fragment shaders.
	if shader.Stage == gcn.GcnShaderStageVertex {
		sampledIndex = 16
		storageIndex = 16
	}

	for _, resource := range resources {
		access, ok := resource.Kind.Access()
		if !ok {
			continue
		}
		var bindingIndex uint32
		switch access {
		case BindingAccessSampledRead:
			bindingIndex = sampledIndex
			sampledIndex++
		case BindingAccessStorageWrite:
			bindingIndex = storageIndex
			storageIndex++
		}
		layout = append(layout, ShaderResourceBinding{
			InstructionOffset: resource.InstructionOffset,
			Kind:              resource.Kind,
			BindingIndex:      bindingIndex,
			Access:            access,
		})
	}

	return layout
}

func StaticResources(layout []ShaderResourceBinding) []SpirvShaderResource {
	resources := make([]SpirvShaderResource, len(layout))
	for i, binding := range layout {
		resources[i] = SpirvShaderResource{
			InstructionOffset: binding.InstructionOffset,
			Kind:              binding.Kind,
		}
	}

	return resources
}
