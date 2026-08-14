package spirv

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

type SpirvShaderStaticLayout map[*gcnSpec.Instruction]ShaderResourceBinding

func BuildStaticLayout(resources []SpirvShaderResource, shader *gcn.GcnShader) SpirvShaderStaticLayout {
	layout := make(map[*gcnSpec.Instruction]ShaderResourceBinding)
	var sampledImageIndex, storageImageIndex uint32
	var translationBufferIndex uint32

	// Offset bindings for vertex shaders so they don't overlap with fragment shaders.
	if shader.Stage == gcn.GcnShaderStageVertex {
		sampledImageIndex = spirvStructs.VertexBindingOffset
		storageImageIndex = spirvStructs.VertexBindingOffset
		translationBufferIndex = spirvStructs.VertexBindingOffset
	}

	for _, resource := range resources {
		var isStorage bool
		var isValid bool
		switch kind := resource.Kind.(type) {
		case ImageAccessKind:
			switch kind {
			case ImageAccessLoad, ImageAccessSample:
				isStorage = false
				isValid = true
			case ImageAccessStore:
				isStorage = true
				isValid = true
			}
		case BufferAccessKind:
			switch kind {
			case BufferAccessLoad:
				isStorage = false
				isValid = true
			case BufferAccessStore:
				isStorage = true
				isValid = true
			}
		case MemoryAccessKind:
			switch kind {
			case MemoryAccessLoad:
				isStorage = false
				isValid = true
			}
		}
		if !isValid {
			continue
		}

		var bindingIndex uint32
		if resource.Type == SpirvShaderResourceTypeImage {
			if isStorage {
				bindingIndex = storageImageIndex
				storageImageIndex++
			} else {
				bindingIndex = sampledImageIndex
				sampledImageIndex++
			}
		} else {
			bindingIndex = translationBufferIndex
			translationBufferIndex++
		}

		layout[resource.Instruction] = ShaderResourceBinding{
			Instruction:  resource.Instruction,
			Type:         resource.Type,
			Kind:         resource.Kind,
			BindingIndex: bindingIndex,
		}
	}

	return layout
}
