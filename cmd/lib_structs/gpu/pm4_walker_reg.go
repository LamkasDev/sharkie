package gpu

import (
	"slices"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func (l *Liverpool) handleSetConfigReg(stream *LiverpoolCommandStream, payload []uint32) {
	l.handleSetRegs(stream, l.Registers.Config[:], "config", gcn.ConfigRegisterNames, payload)
}

func (l *Liverpool) handleSetShaderReg(stream *LiverpoolCommandStream, payload []uint32) {
	l.handleSetRegs(stream, l.Registers.Shader[:], "shader", gcn.ShaderRegisterNames, payload)
}

func (l *Liverpool) handleSetContextReg(stream *LiverpoolCommandStream, payload []uint32) {
	l.handleSetRegs(stream, l.Registers.Context[:], "context", gcn.ContextRegisterNames, payload)
}

func (l *Liverpool) handleSetUserConfigReg(stream *LiverpoolCommandStream, payload []uint32) {
	l.handleSetRegs(stream, l.Registers.UserConfig[:], "user_config", gcn.UserConfigRegisterNames, payload)
}

func (l *Liverpool) handleSetRegsRaw(stream *LiverpoolCommandStream, offset uint32, payload []uint32) {
	switch {
	case offset >= gcn.GcnRegBaseUserConfig:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseUserConfig)
		l.handleSetRegs(stream, l.Registers.UserConfig[:], "user_config", gcn.UserConfigRegisterNames, payload)
	case offset >= gcn.GcnRegBaseContext:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseContext)
		l.handleSetRegs(stream, l.Registers.Context[:], "context", gcn.ContextRegisterNames, payload)
	case offset >= gcn.GcnRegBaseShader:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseShader)
		l.handleSetRegs(stream, l.Registers.Shader[:], "shader", gcn.ShaderRegisterNames, payload)
	case offset >= gcn.GcnRegBaseConfig:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseConfig)
		l.handleSetRegs(stream, l.Registers.Config[:], "config", gcn.ConfigRegisterNames, payload)
	case offset >= gcn.GcnRegBaseSystem:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseSystem)
		l.handleSetRegs(stream, l.Registers.System[:], "system", gcn.SystemRegisterNames, payload)
	}
}

func (l *Liverpool) handleSetRegs(stream *LiverpoolCommandStream, bank []uint32, bankName string, bankRegNames map[uint32]string, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed set regs payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}
	l.StateMutex.Lock()
	offset := payload[0] & 0xFFFF
	for index, value := range payload[1:] {
		bankIndex := int(offset) + index
		if bankIndex < len(bank) {
			bank[bankIndex] = value
			if bankName == "context" && !reg.ContextRegisters[bankIndex] {
				logger.Printf("[%s] UNIMPLEMENTED set %s/%s to %s.\n",
					color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
					color.Blue.Sprint(bankName),
					color.Blue.Sprint(bankRegNames[uint32(bankIndex)]),
					color.Green.Sprintf("0x%X", value),
				)
			}
			if LogPM4Packets {
				logger.Printf("[%s] set %s/%s to %s.\n",
					color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
					color.Blue.Sprint(bankName),
					color.Blue.Sprint(bankRegNames[uint32(bankIndex)]),
					color.Green.Sprintf("0x%X", value),
				)
			}
		}
	}
	l.StateMutex.Unlock()
}

func (l *Liverpool) handleWaitRegMemory(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 6 {
		logger.Printf("[%s] failed wait reg memory payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}

	// Check if we support the memory space.
	function := payload[0] & 0xF
	memorySpace := (payload[0] >> 4) & 0x1
	if memorySpace == 0 {
		// MMIO register poll, skip for now.
		logger.Printf("[%s] failed wait reg memory on mmio register %s.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
			color.Yellow.Sprintf("0x%X", payload[1]),
		)
		return
	}

	// Get address of polled value and the reference value.
	addressLow := uint64(payload[1])
	addressHigh := uint64(payload[2] & 0xFFFF)
	address := uintptr(addressLow | (addressHigh << 32))
	mask := payload[4]
	reference := payload[3] & mask

	// Construct wait reg memory.
	waitRegMem := LiverpoolWaitRegMemory{
		LiverpoolWaitRegMemoryInternal: LiverpoolWaitRegMemoryInternal{
			Function:  function,
			Address:   address,
			Reference: reference,
			Mask:      mask,
		},
	}

	// Add to command stream.
	waitRegMemHash := waitRegMem.Hash()
	waitRegMemIndex, ok := stream.WaitRegMemsMap[waitRegMemHash]
	if !ok {
		waitRegMemIndex = uint32(len(stream.WaitRegMems))
		stream.WaitRegMems = append(stream.WaitRegMems, waitRegMem)
		stream.WaitRegMemsMap[waitRegMemHash] = waitRegMemIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeWaitRegMemory, Index: waitRegMemIndex})

	if LogPM4Packets {
		logger.Printf("[%s] deferred wait on reg memory.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
	}
}

// WaitRegMemCompare evaluates the WAIT_REG_MEM comparison function field.
func WaitRegMemCompare(function, current, reference uint32) bool {
	switch function {
	case 0:
		return true
	case 1:
		return current < reference
	case 2:
		return current <= reference
	case 3:
		return current == reference
	case 4:
		return current != reference
	case 5:
		return current >= reference
	case 6:
		return current > reference
	default:
		return true
	}
}
