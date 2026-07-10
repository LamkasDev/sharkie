package translation

import (
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
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
			t.InvalidateMemory(req.Addr, lib_structs.SystemPageSize)
		} else {
			// logger.Printf("read at 0x%X.\n", req.Addr)
			t.ReadMemory(req.Addr, lib_structs.SystemPageSize)
		}
		structs.CompleteSyncRequest()
	}
}

func (t *GpuTranslator) WriteData(command *gpu.LiverpoolWriteData) {
	if command.Address != 0 {
		dstSlice := unsafe.Slice((*uint32)(unsafe.Pointer(command.Address)), len(command.Data))
		copy(dstSlice, command.Data)
	}
	if logger.LogRenderer {
		logger.Printf("[%s] wrote %s bytes to %s.\n",
			color.Blue.Sprintf("Frame %d", t.currentGuestFrame),
			color.Green.Sprintf("%d", len(command.Data)),
			color.Yellow.Sprintf("0x%X", command.Address),
		)
	}
}

func (t *GpuTranslator) WaitRegMemory(command *gpu.LiverpoolWaitRegMemory) bool {
	if logger.LogRenderer {
		logger.Printf("[%s] waiting on reg memory (address=%s, function=%s, reference=%s).\n",
			color.Blue.Sprintf("Frame %d", t.currentGuestFrame),
			color.Yellow.Sprintf("0x%X", command.Address),
			color.Yellow.Sprintf("0x%X", command.Function),
			color.Yellow.Sprintf("0x%X", command.Reference),
		)
	}
	current := *(*uint32)(unsafe.Pointer(command.Address)) & command.Mask
	if ok := gpu.WaitRegMemCompare(command.Function, current, command.Reference); !ok {
		// TODO: fix this.
		return true
		return false
	}
	if logger.LogRenderer {
		logger.Printf("[%s] finished waiting on reg memory.\n",
			color.Blue.Sprintf("Frame %d", t.currentGuestFrame),
		)
	}

	return true
}
