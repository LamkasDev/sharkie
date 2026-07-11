package gpu

import (
	"slices"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
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

var ImplementedContextRegisters = []string{
	"CB_COLOR_CONTROL",
	"CB_COLOR0_BASE",
	"CB_COLOR0_INFO",
	"CB_COLOR0_PITCH",
	"CB_TARGET_MASK",

	"CB_BLEND0_CONTROL",
	"CB_BLEND_RED", "CB_BLEND_GREEN", "CB_BLEND_BLUE", "CB_BLEND_ALPHA",

	"DB_RENDER_CONTROL",
	"DB_SHADER_CONTROL",
	// "DB_DEPTH_CONTROL",
	// "DB_STENCIL_CONTROL",
	"DB_STENCILREFMASK", "DB_STENCILREFMASK_BF",

	"SPI_PS_INPUT_CNTL_0", "SPI_PS_INPUT_ENA", "SPI_PS_INPUT_ADDR",

	"PA_CL_VTE_CNTL", "PA_SC_MODE_CNTL_0",
	"PA_SC_VPORT_ZMIN_0", "PA_SC_VPORT_ZMAX_0",
	"PA_CL_VPORT_XSCALE", "PA_CL_VPORT_XOFFSET",
	"PA_CL_VPORT_YSCALE", "PA_CL_VPORT_YOFFSET",
	"PA_CL_VPORT_ZSCALE", "PA_CL_VPORT_ZOFFSET",
	"PA_CL_CLIP_CNTL",
	"PA_CL_GB_VERT_CLIP_ADJ", "PA_CL_GB_HORZ_CLIP_ADJ",
	"PA_CL_GB_VERT_DISC_ADJ", "PA_CL_GB_HORZ_DISC_ADJ",

	"PA_SC_SCREEN_SCISSOR_TL", "PA_SC_SCREEN_SCISSOR_BR",
	"PA_SC_VPORT_SCISSOR_0_TL", "PA_SC_VPORT_SCISSOR_0_BR",
	"PA_SC_GENERIC_SCISSOR_TL", "PA_SC_GENERIC_SCISSOR_BR",
	"PA_SC_WINDOW_SCISSOR_TL", "PA_SC_WINDOW_SCISSOR_TL",
	"PA_SC_WINDOW_OFFSET", "PA_SC_WINDOW_OFFSET",
	"PA_SU_SC_MODE_CNTL",

	"PA_SU_HARDWARE_SCREEN_OFFSET",

	"VGT_MULTI_PRIM_IB_RESET_INDX",
	"VGT_MULTI_PRIM_IB_RESET_EN",
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
			if false && bankName == "context" && !slices.Contains(ImplementedContextRegisters, bankRegNames[uint32(bankIndex)]) {
				logger.Printf("[%s] set %s/%s to %s.\n",
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
	waitRegMem := LiverpoolWaitRegMemoryInternal{
		Function:  function,
		Address:   address,
		Reference: reference,
		Mask:      mask,
	}

	// Add to command stream.
	stream.WaitRegMems = append(stream.WaitRegMems, waitRegMem)
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeWaitRegMemory, Index: uint32(len(stream.WaitRegMems) - 1)})

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
