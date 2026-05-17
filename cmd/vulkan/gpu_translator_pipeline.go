package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetPipeline(request GraphicsPipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[request.GraphicsPipelineKey]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createGraphicsPipeline(request)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createGraphicsPipeline: %w", err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[request.GraphicsPipelineKey] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
}

func (t *GpuTranslator) GetComputePipeline(request ComputePipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.computePipelinesMutex.Lock()
	pipeline, ok := t.computePipelines[request.ComputePipelineKey]
	t.computePipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createComputePipeline(request)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createComputePipeline: %w", err)
	}
	t.computePipelinesMutex.Lock()
	t.computePipelines[request.ComputePipelineKey] = pipeline
	t.computePipelinesMutex.Unlock()

	return pipeline, nil
}
