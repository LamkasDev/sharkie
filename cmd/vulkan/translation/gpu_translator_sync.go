package translation

import (
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

func (t *GpuTranslator) memorySyncWorker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	logger.Printf("memory sync worker started.\n")
	for {
		req := structs.WaitForSyncRequest()
		if req.Addr == 0 {
			continue
		}
		if req.IsWrite {
			// logger.Printf("write at 0x%X.\n", req.Addr)
			t.InvalidateMemory(req.Addr, SystemPageSize)
		} else {
			// logger.Printf("read at 0x%X.\n", req.Addr)
			t.ReadMemory(req.Addr, SystemPageSize)
		}
		structs.CompleteSyncRequest()
	}
}

func (t *GpuTranslator) WriteData(command *gpu.LiverpoolWriteData) {
	t.commandBuffer.Writes = append(t.commandBuffer.Writes, command)
}

func (t *GpuTranslator) WaitRegMemory(command *gpu.LiverpoolWaitRegMemory) {
	if command.Satisfied() {
		if logger.LogRendererInternal {
			logger.Printf("[%s] skipped waiting on reg memory (address=%s, function=%s, reference=%s).\n",
				color.Blue.Sprintf("Frame %d", t.currentGuestFrame),
				color.Yellow.Sprintf("0x%X", command.Address),
				color.Yellow.Sprintf("0x%X", command.Function),
				color.Yellow.Sprintf("0x%X", command.Reference),
			)
		}
		return
	}
	t.EndCommandBuffer()
	t.StartCommandBuffer()
	t.commandBuffer.Dependencies = append(t.commandBuffer.Dependencies, command)
}
