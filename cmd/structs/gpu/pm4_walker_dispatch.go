package gpu

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/gookit/color"
)

func (l *Liverpool) handleDispatchDirect(ringName string, payload []uint32) {
	if len(payload) < 3 {
		logger.Printf("[%s] failed dispatch direct payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// End render pass.
	l.StateMutex.Lock()
	defer l.StateMutex.Unlock()

	// Construct dispatch.
	dispatch := LiverpoolDispatch{
		ComputeShader: l.GetShader(GcnShaderStageCompute, l.CsGpuAddress()),
		LiverpoolDispatchInternal: LiverpoolDispatchInternal{
			DimX: payload[0], DimY: payload[1], DimZ: payload[2],
			ThreadX: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_X],
			ThreadY: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Y],
			ThreadZ: l.Registers.Shader[GREG_MM_COMPUTE_NUM_THREAD_Z],

			ComputeShRsrc1: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC1],
			ComputeShRsrc2: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC2],

			UserDataHash: l.SnapshotUserData(),
		},
	}
	dispatch.ComputeShaderAddress = dispatch.ComputeShader.Address

	// Add to command stream.
	l.Stream.Dispatches = append(l.Stream.Dispatches, dispatch)
	l.Stream.Commands = append(l.Stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeDispatch, Index: uint32(len(l.Stream.Dispatches) - 1)})

	if LogPM4Packets {
		logger.Printf("[%s] dispatch direct (dimX=%s, dimY=%s, dimZ=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", dispatch.DimX),
			color.Yellow.Sprintf("0x%X", dispatch.DimY),
			color.Yellow.Sprintf("0x%X", dispatch.DimZ),
		)
	}
}
