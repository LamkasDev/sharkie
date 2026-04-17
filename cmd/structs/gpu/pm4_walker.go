package gpu

import (
	"context"
	"runtime/pprof"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/gookit/color"
)

type PM4Handler func(ringName string, payload []uint32)

const LogPM4Packets = false

func (l *Liverpool) SetupPM4Handlers() {
	l.PM4Handlers[PM4_IT_NOP] = l.handleNop
	l.PM4Handlers[PM4_IT_SET_CONFIG_REG] = l.handleSetConfigReg
	l.PM4Handlers[PM4_IT_SET_SH_REG] = l.handleSetShaderReg
	l.PM4Handlers[PM4_IT_SET_CONTEXT_REG] = l.handleSetContextReg
	l.PM4Handlers[PM4_IT_SET_UCONFIG_REG] = l.handleSetUserConfigReg
	l.PM4Handlers[PM4_IT_WAIT_REG_MEM] = l.handleWaitRegMemory

	l.PM4Handlers[PM4_IT_WRITE_DATA] = l.handleWriteData
	l.PM4Handlers[PM4_WRITE_CONST_RAM] = l.handleWriteConstRam
	l.PM4Handlers[PM4_IT_DMA_DATA] = l.handleDmaData

	l.PM4Handlers[PM4_IT_DRAW_INDEX_AUTO] = l.handleDrawIndexAuto
	l.PM4Handlers[PM4_IT_DRAW_INDEX_2] = l.handleDrawIndex2

	l.PM4Handlers[PM4_IT_CONTEXT_CONTROL] = l.handleContextControl
	l.PM4Handlers[PM4_IT_CLEAR_STATE] = l.handleClearState
	l.PM4Handlers[PM4_ACQUIRE_MEM] = l.handleAcquireMem
	l.PM4Handlers[PM4_IT_NUM_INSTANCES] = l.handleNumInstances
	l.PM4Handlers[PM4_IT_INDEX_TYPE] = l.handleIndexType
	l.PM4Handlers[PM4_IT_INDEX_BUFFER_SIZE] = l.handleIndexBufferSize
	l.PM4Handlers[PM4_IT_EVENT_WRITE_EOP] = l.handleEventWriteEop
	l.PM4Handlers[PM4_IT_EVENT_WRITE_EOS] = l.handleEventWriteEos
	l.PM4Handlers[PM4_IT_WAIT_ON_DE_COUNTER_DIFF] = l.handleWaitOnDeCounterDiff
	l.PM4Handlers[PM4_IT_DISPATCH_DIRECT] = l.handleDispatchDirect
}

// Walk drains both the graphics and compute rings, decoding every PM4 packet and updating GPU register state.
func (l *Liverpool) Walk() {
	asm.GCFence.Store(true)

	l.RingMutex.Lock()
	defer l.RingMutex.Unlock()

	var wg sync.WaitGroup
	wg.Add(2)
	go pprof.Do(context.Background(), pprof.Labels("name", "WalkGraphicsBuffer"), func(ctx context.Context) {
		for i, buffer := range l.GraphicsRing.Pending {
			logger.Printf("[%s] walking graphics pm4 buffer %s (length=%s).\n",
				color.Green.Sprint("PM4"),
				color.Green.Sprintf("%d", i),
				color.Green.Sprintf("%d", buffer.SizeDW),
			)
			l.walkIndirectBuffer("GFX", buffer)
		}
		l.GraphicsRing.Pending = l.GraphicsRing.Pending[:0]
		wg.Done()
	})
	go pprof.Do(context.Background(), pprof.Labels("name", "WalkComputeBuffer"), func(ctx context.Context) {
		for i, buffer := range l.ComputeRing.Pending {
			logger.Printf("[%s] walking compute pm4 buffer %s (length=%s).\n",
				color.Green.Sprint("PM4"),
				color.Green.Sprintf("%d", i),
				color.Green.Sprintf("%d", buffer.SizeDW),
			)
			l.walkIndirectBuffer("COM", buffer)
		}
		l.ComputeRing.Pending = l.ComputeRing.Pending[:0]
		wg.Done()
	})
	wg.Wait()
	logger.Printf(
		"[%s] finished walking pm4 buffers.\n",
		color.Green.Sprint("PM4"),
	)

	asm.GCFence.Store(false)
}

func (l *Liverpool) walkIndirectBuffer(ringName string, buffer PM4IndirectBuffer) {
	if buffer.Address == 0 || buffer.SizeDW == 0 {
		return
	}

	dwords := unsafe.Slice((*uint32)(unsafe.Pointer(buffer.Address)), int(buffer.SizeDW))
	l.walkStream(ringName, dwords)
}

func (l *Liverpool) walkStream(ringName string, dwords []uint32) {
	i := 0
	for i < len(dwords) {
		// Type-2 is the single DWORD NOP padding.
		header := dwords[i]
		if header == 0 || header == PM4_HEADER_TYPE2 {
			i++
			continue
		}

		// Extract header data.
		headerType := (header >> 30) & 0x3
		count := int((header>>16)&0x3FFF) + 1
		opcode := uint8((header >> 8) & 0xFF)
		end := i + 1 + count

		// Check if the packet is truncated.
		if end > len(dwords) {
			logger.Printf("[%s] truncated %s-pm4 opcode %s (expected=%s, got=%s).\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Green.Sprintf("%d", headerType),
				color.Yellow.Sprintf("0x%X", opcode),
				color.Green.Sprintf("%d", count),
				color.Green.Sprintf("%d", len(dwords)-i-1),
			)
			break
		}

		switch headerType {
		case PM4_TYPE_0:
			regOffset := header & 0xFFFF
			l.handleSetRegsRaw(ringName, regOffset, dwords[i+1:end])
		case PM4_TYPE_3:
			l.dispatchType3Packet(ringName, opcode, dwords[i+1:end])
		}

		i = end
	}
}

func (l *Liverpool) dispatchType3Packet(ringName string, opcode uint8, payload []uint32) {
	if handler, ok := l.PM4Handlers[opcode]; ok {
		handler(ringName, payload)
		return
	}

	logger.Printf("[%s] unknown pm4 opcode %s.\n",
		color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		color.Yellow.Sprintf("0x%X", opcode),
	)
}

func (l *Liverpool) handleNop(ringName string, payload []uint32) {}

func (l *Liverpool) handleContextControl(ringName string, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed context control payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	loadControl := payload[0]
	shadowControl := payload[1]
	if LogPM4Packets {
		logger.Printf("[%s] attempted context switch (load=%s, shadow=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", loadControl),
			color.Yellow.Sprintf("0x%X", shadowControl),
		)
	}
}

func (l *Liverpool) handleClearState(ringName string, payload []uint32) {
	l.StateMutex.Lock()
	for i := range l.Registers.Context {
		l.Registers.Context[i] = 0
	}
	for i := range l.Registers.Shader {
		l.Registers.Shader[i] = 0
	}
	l.DrawState = LiverpoolDrawState{}
	l.StateMutex.Unlock()
	if LogPM4Packets {
		logger.Printf("[%s] cleared state.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
	}
}

func (l *Liverpool) handleAcquireMem(ringName string, payload []uint32) {
	if len(payload) < 6 {
		logger.Printf("[%s] failed acquire mem payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	if LogPM4Packets {
		logger.Printf("[%s] attempted acquire mem (payload[0]=%s, payload[1]=%s, payload[2]=%s, payload[3]=%s, payload[4]=%s, payload[5]=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", payload[0]),
			color.Yellow.Sprintf("0x%X", payload[1]),
			color.Yellow.Sprintf("0x%X", payload[2]),
			color.Yellow.Sprintf("0x%X", payload[3]),
			color.Yellow.Sprintf("0x%X", payload[4]),
			color.Yellow.Sprintf("0x%X", payload[5]),
		)
	}
}

func (l *Liverpool) handleNumInstances(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed num instances payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.DrawState.InstanceCount = payload[0]
	if LogPM4Packets {
		logger.Printf("[%s] set num instances to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.InstanceCount),
		)
	}
}

func (l *Liverpool) handleIndexType(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed index type payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.DrawState.IndexType = payload[0] & 1
	if LogPM4Packets {
		switch l.DrawState.IndexType {
		case 0:
			logger.Printf("[%s] set index type to 16-bit.\n", color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)))
		case 1:
			logger.Printf("[%s] set index type to 32-bit.\n", color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)))
		}
	}
}

func (l *Liverpool) handleIndexBufferSize(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed index buffer size payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.DrawState.IndexBufferSize = payload[0]
	if LogPM4Packets {
		logger.Printf("[%s] set index buffer size to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.IndexBufferSize),
		)
	}
}

func (l *Liverpool) handleEventWriteEop(ringName string, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] failed event write eop payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	dataHigh := uint32(0)
	if len(payload) >= 5 {
		dataHigh = payload[4]
	}
	l.handleEventWriteEopEos(ringName, "eop", payload[1], payload[2], payload[3], dataHigh)
}

func (l *Liverpool) handleEventWriteEos(ringName string, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] failed event write eos payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.handleEventWriteEopEos(ringName, "eos", payload[1], payload[2], payload[3], 0)
}

func (l *Liverpool) handleEventWriteEopEos(ringName, kind string, addrLow, addrHighAndSel, dataLow, dataHigh uint32) {
	// Get address of destination.
	addressLow := uint64(addrLow)
	addressHigh := uint64(addrHighAndSel & 0xFFFF)
	address := uintptr(addressLow | (addressHigh << 32))
	if address == 0 {
		logger.Printf("[%s] failed write %s data invalid address.\n",
			color.Green.Sprintf("PM4-%s", ringName),
			color.Blue.Sprint(kind),
		)
		return
	}

	// Write data.
	dataSelection := (addrHighAndSel >> 29) & 0x7
	switch dataSelection {
	case 0: // No write.
		if LogPM4Packets {
			logger.Printf("[%s] skipped %s write to %s.\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Blue.Sprint(kind),
				color.Yellow.Sprintf("0x%X", address),
			)
		}
	case 1: // 32-bit value.
		*(*uint32)(unsafe.Pointer(address)) = dataLow
		if LogPM4Packets {
			logger.Printf("[%s] wrote %s 32-bit %s to %s.\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Blue.Sprint(kind),
				color.Yellow.Sprintf("0x%X", dataLow),
				color.Yellow.Sprintf("0x%X", address),
			)
		}
	case 2: // 64-bit value.
		value := uint64(dataLow) | uint64(dataHigh)<<32
		*(*uint64)(unsafe.Pointer(address)) = value
		if LogPM4Packets {
			logger.Printf("[%s] wrote %s 64-bit %s to %s.\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Blue.Sprint(kind),
				color.Yellow.Sprintf("0x%X", value),
				color.Yellow.Sprintf("0x%X", address),
			)
		}
	case 3: // GPU timestamp.
		if LogPM4Packets {
			logger.Printf("[%s] wrote %s GPU timestamp to %s.\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Blue.Sprint(kind),
				color.Yellow.Sprintf("0x%X", address),
			)
		}
		*(*uint64)(unsafe.Pointer(address)) = 0
	}
}

func (l *Liverpool) handleWaitOnDeCounterDiff(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed wait on de counter payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	// diff := payload[0] & 0xFF
	// TODO: this
}

func (l *Liverpool) handleDispatchDirect(ringName string, payload []uint32) {
	if len(payload) < 3 {
		logger.Printf("[%s] failed dispatch direct payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	l.StateMutex.Lock()
	computeShPgmLo := l.Registers.Shader[GREG_MM_COMPUTE_PGM_LO]
	computeShPgmHi := l.Registers.Shader[GREG_MM_COMPUTE_PGM_HI]
	csAddress := (uintptr(computeShPgmLo) | uintptr(computeShPgmHi)<<32) << 8
	l.StateMutex.Unlock()

	// Force it to load.
	l.GetShader(GcnShaderStageCompute, csAddress)

	if LogPM4Packets {
		logger.Printf("[%s] dispatch direct (payload[0]=%s, payload[1]=%s, payload[2]=%s, payload[3]=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", payload[0]),
			color.Yellow.Sprintf("0x%X", payload[1]),
			color.Yellow.Sprintf("0x%X", payload[2]),
			color.Yellow.Sprintf("0x%X", payload[3]),
		)
	}
}
