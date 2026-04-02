package gpu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type PM4Handler func(payload []uint32)

func (l *Liverpool) SetupPM4Handlers() {
	l.PM4Handlers[PM4_IT_NOP] = l.handleNop
	l.PM4Handlers[PM4_IT_SET_SH_REG] = l.handleSetShReg
	l.PM4Handlers[PM4_IT_SET_CONTEXT_REG] = l.handleSetContextReg
}

// Walk drains both the graphics and compute rings, decoding every PM4 packet and updating GPU register state.
func (l *Liverpool) Walk() {
	for _, buffer := range l.GraphicsRing.Pending {
		l.walkIndirectBuffer(buffer)
	}
	l.GraphicsRing.Pending = l.GraphicsRing.Pending[:0]

	for _, buffer := range l.ComputeRing.Pending {
		l.walkIndirectBuffer(buffer)
	}
	l.ComputeRing.Pending = l.ComputeRing.Pending[:0]
}

func (l *Liverpool) walkIndirectBuffer(buffer PM4IndirectBuffer) {
	address := uintptr(buffer.AddressLow) | (uintptr(buffer.AddressHigh) << 32)
	if address == 0 || buffer.SizeDW == 0 {
		return
	}

	dwords := unsafe.Slice((*uint32)(unsafe.Pointer(address)), int(buffer.SizeDW))
	l.walkStream(dwords)
}

func (l *Liverpool) walkStream(dwords []uint32) {
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
			logger.Printf("truncated pm4 opcode %s (expected=%s, got=%s).\n",
				color.Yellow.Sprintf("0x%X", opcode),
				color.Green.Sprintf("%d", count),
				color.Green.Sprintf("%d", len(dwords)-i-1),
			)
			break
		}

		l.dispatchPacket(opcode, dwords[i+1:end])
		i = end
	}
}

func (l *Liverpool) dispatchPacket(opcode uint8, payload []uint32) {
	if handler, ok := l.PM4Handlers[opcode]; ok {
		handler(payload)
		return
	}

	logger.Printf("unknown pm4 opcode %s (len=%s).\n",
		color.Yellow.Sprintf("0x%X", opcode),
		color.Green.Sprintf("%d", len(payload)),
	)
}

func (l *Liverpool) handleNop(payload []uint32) {}

func (l *Liverpool) handleSetShReg(payload []uint32) {
	l.handleSetRegs(l.Registers.Shader[:], ShaderRegisterNames, "shader", payload)
}

func (l *Liverpool) handleSetContextReg(payload []uint32) {
	l.handleSetRegs(l.Registers.Context[:], ContextRegisterNames, "context", payload)
}

func (l *Liverpool) handleSetRegs(bank []uint32, regNames map[uint32]string, bankName string, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("set regs payload too short.\n")
		return
	}
	offset := payload[0] & 0xFFFF
	for index, value := range payload[1:] {
		bankIndex := int(offset) + index
		if bankIndex < len(bank) {
			bank[bankIndex] = value
			logger.Printf("set %s (%s/%s) to %s.\n",
				color.Blue.Sprint(regNames[uint32(bankIndex)]),
				color.Blue.Sprint(bankName),
				color.Yellow.Sprintf("0x%X", bankIndex),
				color.Green.Sprintf("%d", value),
			)
		}
	}
}
