package gpu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type PM4Handler func(stream *LiverpoolCommandStream, payload []uint32)

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
func (l *Liverpool) Walk() []*LiverpoolCommandStream {
	asm.GCFence.Store(true)

	l.RingMutex.Lock()
	pendingGraphics := l.GraphicsRing.Pending
	l.GraphicsRing.Pending = l.GraphicsRing.Pending[:0]
	pendingCompute := l.ComputeRing.Pending
	l.ComputeRing.Pending = l.ComputeRing.Pending[:0]
	l.RingMutex.Unlock()

	graphicsStream := NewLiverpoolCommandStream("GFX")
	for i, buffer := range pendingGraphics {
		logger.Printf("[%s] walking GFX pm4 buffer %s (length=%s).\n",
			color.Green.Sprint("PM4"),
			color.Green.Sprintf("%d", i),
			color.Green.Sprintf("%d", buffer.SizeDW),
		)
		l.walkIndirectBuffer(graphicsStream, buffer)
	}
	computeStream := NewLiverpoolCommandStream("COM")
	for i, buffer := range pendingCompute {
		logger.Printf("[%s] walking COM pm4 buffer %s (length=%s).\n",
			color.Green.Sprint("PM4"),
			color.Green.Sprintf("%d", i),
			color.Green.Sprintf("%d", buffer.SizeDW),
		)
		l.walkIndirectBuffer(computeStream, buffer)
	}
	logger.Printf(
		"[%s] finished walking pm4 buffers.\n",
		color.Green.Sprint("PM4"),
	)

	asm.GCFence.Store(false)

	return []*LiverpoolCommandStream{computeStream, graphicsStream}
}

func (l *Liverpool) walkIndirectBuffer(stream *LiverpoolCommandStream, buffer PM4IndirectBuffer) {
	if buffer.Address == 0 || buffer.SizeDW == 0 {
		return
	}

	dwords := unsafe.Slice((*uint32)(unsafe.Pointer(buffer.Address)), int(buffer.SizeDW))
	l.walkStream(stream, dwords)
}

func (l *Liverpool) walkStream(stream *LiverpoolCommandStream, dwords []uint32) {
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
				color.Green.Sprintf("PM4-%s", stream.Name),
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
			l.handleSetRegsRaw(stream, regOffset, dwords[i+1:end])
		case PM4_TYPE_3:
			l.dispatchType3Packet(stream, opcode, dwords[i+1:end])
		}

		i = end
	}
}

func (l *Liverpool) dispatchType3Packet(stream *LiverpoolCommandStream, opcode uint8, payload []uint32) {
	if handler, ok := l.PM4Handlers[opcode]; ok {
		handler(stream, payload)
		return
	}

	logger.Printf("[%s] unknown pm4 opcode %s.\n",
		color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		color.Yellow.Sprintf("0x%X", opcode),
	)
}

func (l *Liverpool) handleNop(stream *LiverpoolCommandStream, payload []uint32) {}

func (l *Liverpool) handleContextControl(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed context control payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	loadControl := payload[0]
	shadowControl := payload[1]
	if LogPM4Packets {
		logger.Printf("[%s] attempted context switch (load=%s, shadow=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", loadControl),
			color.Yellow.Sprintf("0x%X", shadowControl),
		)
	}
}

func (l *Liverpool) handleClearState(stream *LiverpoolCommandStream, payload []uint32) {
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
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
	}
}

func (l *Liverpool) handleAcquireMem(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 6 {
		logger.Printf("[%s] failed acquire mem payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	if LogPM4Packets {
		logger.Printf("[%s] attempted acquire mem (payload[0]=%s, payload[1]=%s, payload[2]=%s, payload[3]=%s, payload[4]=%s, payload[5]=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", payload[0]),
			color.Yellow.Sprintf("0x%X", payload[1]),
			color.Yellow.Sprintf("0x%X", payload[2]),
			color.Yellow.Sprintf("0x%X", payload[3]),
			color.Yellow.Sprintf("0x%X", payload[4]),
			color.Yellow.Sprintf("0x%X", payload[5]),
		)
	}
}

func (l *Liverpool) handleNumInstances(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed num instances payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	l.DrawState.InstanceCount = payload[0]
	if LogPM4Packets {
		logger.Printf("[%s] set num instances to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.InstanceCount),
		)
	}
}

func (l *Liverpool) handleIndexType(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed index type payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	l.DrawState.IndexType = payload[0] & 1
	if LogPM4Packets {
		switch l.DrawState.IndexType {
		case 0:
			logger.Printf("[%s] set index type to 16-bit.\n", color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)))
		case 1:
			logger.Printf("[%s] set index type to 32-bit.\n", color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)))
		}
	}
}

func (l *Liverpool) handleIndexBufferSize(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed index buffer size payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	l.DrawState.IndexBufferSize = payload[0]
	if LogPM4Packets {
		logger.Printf("[%s] set index buffer size to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", l.DrawState.IndexBufferSize),
		)
	}
}

func (l *Liverpool) handleEventWriteEop(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] failed event write eop payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	dataHigh := uint32(0)
	if len(payload) >= 5 {
		dataHigh = payload[4]
	}
	l.handleEventWriteEopEos(stream, "eop", payload[1], payload[2], payload[3], dataHigh)
}

func (l *Liverpool) handleEventWriteEos(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] failed event write eos payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	l.handleEventWriteEopEos(stream, "eos", payload[1], payload[2], payload[3], 0)
}

func (l *Liverpool) handleEventWriteEopEos(stream *LiverpoolCommandStream, kind string, addrLow, addrHighAndSel, dataLow, dataHigh uint32) {
	// Get address of destination.
	addressLow := uint64(addrLow)
	addressHigh := uint64(addrHighAndSel & 0xFFFF)
	address := uintptr(addressLow | (addressHigh << 32))
	if address == 0 {
		logger.Printf("[%s] failed write %s data invalid address.\n",
			color.Green.Sprintf("PM4-%s", stream.Name),
			color.Blue.Sprint(kind),
		)
		return
	}

	// Write data.
	dataSelection := (addrHighAndSel >> 29) & 0x7
	var data []uint32
	switch dataSelection {
	case 0: // No write.
		return
	case 1: // 32-bit value.
		data = []uint32{dataLow}
	case 2: // 64-bit value.
		data = []uint32{dataLow, dataHigh}
	case 3: // GPU timestamp.
		data = []uint32{0, 0}
	}

	// Construct write data.
	writeData := LiverpoolWriteData{
		Address: address,
		Data:    data,
	}

	// Add to command stream.
	stream.WriteDatas = append(stream.WriteDatas, writeData)
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeWriteData, Index: uint32(len(stream.WriteDatas) - 1)})

	if LogPM4Packets {
		logger.Printf("[%s] deferred writing %s data to %s.\n",
			color.Green.Sprintf("PM4-%s", stream.Name),
			color.Blue.Sprint(kind),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
}

func (l *Liverpool) handleWaitOnDeCounterDiff(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] failed wait on de counter payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	// diff := payload[0] & 0xFF
	// TODO: this
}
