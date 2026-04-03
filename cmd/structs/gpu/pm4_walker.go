package gpu

import (
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
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

	l.PM4Handlers[PM4_IT_CONTEXT_CONTROL] = l.handleContextControl
	l.PM4Handlers[PM4_IT_CLEAR_STATE] = l.handleClearState
	l.PM4Handlers[PM4_ACQUIRE_MEM] = l.handleAcquireMem
	l.PM4Handlers[PM4_IT_NUM_INSTANCES] = l.handleNumInstances
	l.PM4Handlers[PM4_IT_INDEX_BUFFER_SIZE] = l.handleIndexBufferSize
	l.PM4Handlers[PM4_IT_WRITE_DATA] = l.handleWriteData
	l.PM4Handlers[PM4_IT_WAIT_REG_MEM] = l.handleWaitRegMemory
	l.PM4Handlers[PM4_WRITE_CONST_RAM] = l.handleWriteConstRam
}

// Walk drains both the graphics and compute rings, decoding every PM4 packet and updating GPU register state.
func (l *Liverpool) Walk() {
	l.RingMutex.Lock()
	var wg sync.WaitGroup
	wg.Go(func() {
		for i, buffer := range l.GraphicsRing.Pending {
			logger.Printf("[%s] walking graphics pm4 buffer %s (length=%s).\n",
				color.Green.Sprint("PM4"),
				color.Green.Sprintf("%d", i),
				color.Green.Sprintf("%d", buffer.SizeDW),
			)
			l.walkIndirectBuffer("GFX", buffer)
		}
		l.GraphicsRing.Pending = l.GraphicsRing.Pending[:0]
	})
	wg.Go(func() {
		for i, buffer := range l.ComputeRing.Pending {
			logger.Printf("[%s] walking compute pm4 buffer %s (length=%s).\n",
				color.Green.Sprint("PM4"),
				color.Green.Sprintf("%d", i),
				color.Green.Sprintf("%d", buffer.SizeDW),
			)
			l.walkIndirectBuffer("COM", buffer)
		}
		l.ComputeRing.Pending = l.ComputeRing.Pending[:0]
	})
	wg.Wait()
	logger.Printf(
		"[%s] finished walking pm4 buffers.\n",
		color.Green.Sprint("PM4"),
	)
	l.RingMutex.Unlock()
}

func (l *Liverpool) walkIndirectBuffer(ringName string, buffer PM4IndirectBuffer) {
	address := uintptr(buffer.AddressLow) | (uintptr(buffer.AddressHigh) << 32)
	if address == 0 || buffer.SizeDW == 0 {
		return
	}

	dwords := unsafe.Slice((*uint32)(unsafe.Pointer(address)), int(buffer.SizeDW))
	l.walkStream(ringName, dwords)
}

func (l *Liverpool) walkStream(ringName string, dwords []uint32) {
	i := 0
	for i < len(dwords) {
		header := dwords[i]

		// Type-2 is the single DWORD NOP padding.
		if header == PM4_HEADER_TYPE2 {
			i++
			continue
		}

		// All normal packets must be Type-3.
		if (header>>30)&3 != PM4_TYPE_3 {
			i++
			continue
		}

		// count is the number of payload DWORDs.
		count := int((header>>16)&0x3FFF) + 1
		opcode := uint8((header >> 8) & 0xFF)

		end := i + 1 + count
		if end > len(dwords) {
			logger.Printf("[%s] truncated pm4 opcode %s (expected=%s, got=%s).\n",
				color.Green.Sprintf("PM4-%s", ringName),
				color.Yellow.Sprintf("0x%X", opcode),
				color.Green.Sprintf("%d", count),
				color.Green.Sprintf("%d", len(dwords)-i-1),
			)
			break
		}

		l.dispatchPacket(ringName, opcode, dwords[i+1:end])
		i = end
	}
}

func (l *Liverpool) dispatchPacket(ringName string, opcode uint8, payload []uint32) {
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
		logger.Printf("[%s] context control payload too short.\n",
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
		logger.Printf("[%s] acquire mem payload too short.\n",
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
		logger.Printf("[%s] num instances payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.StateMutex.Lock()
	l.DrawState.InstanceCount = payload[0]
	l.StateMutex.Unlock()
	if LogPM4Packets {
		logger.Printf("[%s] set num instances to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.InstanceCount),
		)
	}
}

func (l *Liverpool) handleIndexBufferSize(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] index buffer size payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.StateMutex.Lock()
	l.DrawState.IndexBufferSize = payload[0]
	l.StateMutex.Unlock()
	if LogPM4Packets {
		logger.Printf("[%s] set index buffer size to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.IndexBufferSize),
		)
	}
}

func (l *Liverpool) handleWriteData(ringName string, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] write data payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Check if we support the write destination.
	destSelection := (payload[0] >> 8) & 0x7
	switch destSelection {
	case 0, 1, 5:
	default:
		logger.Printf("[%s] write data on non-memory destination %s skipped.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", destSelection),
		)
		return
	}

	// Get address of destination.
	addressLow := uint64(payload[1])
	addressHigh := uint64(payload[2] & 0xFFFF)
	address := uintptr(addressLow | (addressHigh << 32))
	if address == 0 {
		return
	}

	// Write data.
	data := payload[3:]
	dstSlice := unsafe.Slice((*uint32)(unsafe.Pointer(address)), len(data))
	copy(dstSlice, data)

	if LogPM4Packets {
		logger.Printf("[%s] wrote %s bytes to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", len(data)),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
}

func (l *Liverpool) handleWriteConstRam(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] write const ram payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	offset := int(payload[0] & 0xFFFF)
	data := payload[1:]
	if offset+len(data) > LiverpoolConstRamSize {
		logger.Printf("[%s] failed write const ram outside bounds (offset=%s, size=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", offset),
			color.Green.Sprintf("%d", len(data)),
		)
		return
	}

	// Write data.
	l.StateMutex.Lock()
	copy(l.ConstRam[offset:], data)
	l.StateMutex.Unlock()

	if LogPM4Packets {
		logger.Printf("[%s] wrote %s bytes to const ram at %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", len(data)),
			color.Yellow.Sprintf("0x%X", offset),
		)
	}
}
