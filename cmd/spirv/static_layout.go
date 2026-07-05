package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// BuildStaticLayout assigns Set-2 array indices per access kind (sampled and storage
// each have their own binding in the pipeline layout).
func BuildStaticLayout(resources []SpirvShaderResource) []ShaderResourceBinding {
	var layout []ShaderResourceBinding
	var sampledIndex, storageIndex uint32
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

// SupportsStaticBindings determines if a shader supports static bindings (we can resolve descriptors at runtime).
func SupportsStaticBindings(shader *gcn.GcnShader, layout []ShaderResourceBinding) bool {
	switch shader.Address {
	case 0xFE6DD1100, 0xFE6DD1400: // Compute
		return true
	case 0xFE6DC9A00, 0xFE6DC9700, 0xFE6DCA000, 0xFE6DCCE00, 0xFE6DD0200: // Fragment
		return true
	default:
		return false
	}
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
