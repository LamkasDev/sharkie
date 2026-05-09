package vulkan

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) dispatchCompute(frame uint64, commandBuffer vk.CommandBuffer, dispatch *LiverpoolComputeDispatch) {
	// Get shader module.
	csSpirv := t.GetShader(dispatch.ComputeShader)
	csModule, err := t.GetShaderModule(csSpirv)
	if err != nil {
		return
	}

	_ = csModule
}
