package gpu

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func (l *Liverpool) handleDrawIndexAuto(ringName string, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed draw index auto payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Record draw.
	count := payload[0]
	drawCall := l.NewDrawCall(count, false)
	l.StateMutex.Lock()
	l.PendingDrawCalls = append(l.PendingDrawCalls, drawCall)
	l.StateMutex.Unlock()

	if LogPM4Packets {
		logger.Printf("[%s] draw index auto (vertex=%s, prim=%s, rt=%s, vs=%s, ps=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", count),
			color.Green.Sprintf("%d", drawCall.PrimType),
			color.Yellow.Sprintf("0x%X", drawCall.RtGpuAddress()),
			color.Yellow.Sprintf("0x%X", drawCall.VsGpuAddress()),
			color.Yellow.Sprintf("0x%X", drawCall.PsGpuAddress()),
		)
	}
}

func (l *Liverpool) handleDrawIndex2(ringName string, payload []uint32) {
	if len(payload) < 5 {
		logger.Printf("[%s] failed draw index 2 payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Record draw.
	l.DrawState.IndexBase = uintptr(uint64(payload[1]) | uint64(payload[2])<<32)
	l.DrawState.IndexBufferSize = payload[0]
	count := payload[3]
	drawCall := l.NewDrawCall(count, true)
	l.StateMutex.Lock()
	l.PendingDrawCalls = append(l.PendingDrawCalls, drawCall)
	l.StateMutex.Unlock()

	if LogPM4Packets {
		logger.Printf("[%s] draw index 2 (index_count=%s, index_base=%s, prim=%s, rt=%s, vs=%s, ps=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", count),
			color.Yellow.Sprintf("0x%X", drawCall.IndexBase),
			color.Green.Sprintf("%d", drawCall.PrimType),
			color.Yellow.Sprintf("0x%X", drawCall.RtGpuAddress()),
			color.Yellow.Sprintf("0x%X", drawCall.VsGpuAddress()),
			color.Yellow.Sprintf("0x%X", drawCall.PsGpuAddress()),
		)
	}
}
