package gpu

import (
	"runtime"
	"slices"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/gookit/color"
)

func (l *Liverpool) handleSetConfigReg(ringName string, payload []uint32) {
	l.handleSetRegs(ringName, l.Registers.Config[:], "config", gcn.ConfigRegisterNames, payload)
}

func (l *Liverpool) handleSetShaderReg(ringName string, payload []uint32) {
	l.handleSetRegs(ringName, l.Registers.Shader[:], "shader", gcn.ShaderRegisterNames, payload)
}

func (l *Liverpool) handleSetContextReg(ringName string, payload []uint32) {
	l.handleSetRegs(ringName, l.Registers.Context[:], "context", gcn.ContextRegisterNames, payload)
}

func (l *Liverpool) handleSetUserConfigReg(ringName string, payload []uint32) {
	l.handleSetRegs(ringName, l.Registers.UserConfig[:], "user_config", gcn.UserConfigRegisterNames, payload)
}

func (l *Liverpool) handleSetRegsRaw(ringName string, offset uint32, payload []uint32) {
	switch {
	case offset >= gcn.GcnRegBaseUserConfig:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseUserConfig)
		l.handleSetRegs(ringName, l.Registers.UserConfig[:], "user_config", gcn.UserConfigRegisterNames, payload)
	case offset >= gcn.GcnRegBaseContext:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseContext)
		l.handleSetRegs(ringName, l.Registers.Context[:], "context", gcn.ContextRegisterNames, payload)
	case offset >= gcn.GcnRegBaseShader:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseShader)
		l.handleSetRegs(ringName, l.Registers.Shader[:], "shader", gcn.ShaderRegisterNames, payload)
	case offset >= gcn.GcnRegBaseConfig:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseConfig)
		l.handleSetRegs(ringName, l.Registers.Config[:], "config", gcn.ConfigRegisterNames, payload)
	case offset >= gcn.GcnRegBaseSystem:
		payload = slices.Insert(payload, 0, offset-gcn.GcnRegBaseSystem)
		l.handleSetRegs(ringName, l.Registers.System[:], "system", gcn.SystemRegisterNames, payload)
	}
}

func (l *Liverpool) handleSetRegs(ringName string, bank []uint32, bankName string, bankRegNames map[uint32]string, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed set regs payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	l.StateMutex.Lock()
	offset := payload[0] & 0xFFFF
	for index, value := range payload[1:] {
		bankIndex := int(offset) + index
		if bankIndex < len(bank) {
			bank[bankIndex] = value
			if LogPM4Packets {
				logger.Printf("[%s] set %s/%s to %s.\n",
					color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
					color.Blue.Sprint(bankName),
					color.Blue.Sprint(bankRegNames[uint32(bankIndex)]),
					color.Green.Sprintf("0x%X", value),
				)
			}
		}
	}
	l.StateMutex.Unlock()
}

func (l *Liverpool) handleWaitRegMemory(ringName string, payload []uint32) {
	if len(payload) < 6 {
		logger.Printf("[%s] failed wait reg memory payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Check if we support the memory space.
	function := payload[0] & 0xF
	memorySpace := (payload[0] >> 4) & 0x1
	if memorySpace == 0 {
		// MMIO register poll, skip for now.
		logger.Printf("[%s] failed wait reg memory on mmio register %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
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

	// Compare the values.
	if LogPM4Packets {
		logger.Printf("[%s] waiting on reg memory (address=%s, function=%s, reference=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", address),
			color.Yellow.Sprintf("0x%X", function),
			color.Yellow.Sprintf("0x%X", reference),
		)
	}
	for {
		var current uint32
		if address != 0 {
			current = *(*uint32)(unsafe.Pointer(address)) & mask
		}
		if ok := waitRegMemCompare(function, current, reference); !ok {
			runtime.Gosched()
		}
		break
	}
	if LogPM4Packets {
		logger.Printf("[%s] finished wait on reg memory.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
	}
}

// waitRegMemCompare evaluates the WAIT_REG_MEM comparison function field.
func waitRegMemCompare(function, current, reference uint32) bool {
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
