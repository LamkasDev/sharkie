package gpu

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/gookit/color"
)

func (l *Liverpool) handleDispatchDirect(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 3 {
		logger.Printf("[%s] failed dispatch direct payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}

	// Lock registers.
	l.StateMutex.Lock()
	defer l.StateMutex.Unlock()

	// Construct resource binds.
	bindResources := LiverpoolBindResources{
		LiverpoolBindResourcesInternal: LiverpoolBindResourcesInternal{
			UserDataHash: l.SnapshotUserData(),
		},
	}
	if address := l.CsGpuAddress(); address != 0 {
		bindResources.ComputeShader = l.GetShader(GcnShaderStageCompute, address)
		bindResources.ComputeShaderAddress = address
		bindResources.ComputeContext = common.SpirvComputeShaderContext{
			ThreadX: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_X],
			ThreadY: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Y],
			ThreadZ: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Z],
		}
	} else {
		panic("no compute shader")
	}

	// Add to command stream.
	resHash := bindResources.Hash()
	resIndex, ok := stream.BindResourcesMap[resHash]
	if !ok {
		resIndex = uint32(len(stream.BindResources))
		stream.BindResources = append(stream.BindResources, bindResources)
		stream.BindResourcesMap[resHash] = resIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeBindResources, Index: resIndex})

	// Construct pipeline state.
	bindPipeline := LiverpoolBindComputePipeline{
		LiverpoolBindComputePipelineInternal: LiverpoolBindComputePipelineInternal{},
	}

	// Add to command stream.
	bindHash := bindPipeline.Hash()
	bindIndex, ok := stream.ComputePipelinesMap[bindHash]
	if !ok {
		bindIndex = uint32(len(stream.ComputePipelines))
		stream.ComputePipelines = append(stream.ComputePipelines, bindPipeline)
		stream.ComputePipelinesMap[bindHash] = bindIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeBindComputePipeline, Index: bindIndex})

	// Construct dispatch.
	dispatch := LiverpoolDispatch{
		LiverpoolDispatchInternal: LiverpoolDispatchInternal{
			DimX: payload[0], DimY: payload[1], DimZ: payload[2],
			ThreadX: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_X],
			ThreadY: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Y],
			ThreadZ: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Z],

			ComputeShRsrc1: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC1],
			ComputeShRsrc2: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC2],

			UserDataHash: bindResources.UserDataHash,
		},
	}

	// Add to command stream.
	disHash := dispatch.Hash()
	disIndex, ok := stream.DispatchesMap[disHash]
	if !ok {
		disIndex = uint32(len(stream.DispatchesMap))
		stream.Dispatches = append(stream.Dispatches, dispatch)
		stream.DispatchesMap[disHash] = disIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeDispatch, Index: disIndex})

	if LogPM4Packets {
		logger.Printf("[%s] dispatch direct (dimX=%s, dimY=%s, dimZ=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", dispatch.DimX),
			color.Yellow.Sprintf("0x%X", dispatch.DimY),
			color.Yellow.Sprintf("0x%X", dispatch.DimZ),
		)
	}
}
