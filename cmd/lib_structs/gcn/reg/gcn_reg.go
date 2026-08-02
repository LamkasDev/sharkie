package reg

import "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"

type Reg uint32

func (r Reg) Extract(shift, mask uint32) uint32 {
	return (uint32(r) >> shift) & mask
}

func (r Reg) ExtractBool(shift uint32) bool {
	return ((uint32(r) >> shift) & 1) != 0
}

// Fast O(1) array lookup for context registers.
var ContextRegisters [0x10000]bool

func init() {
	ContextRegisters[gcn.GREG_MM_CB_COLOR_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_BASE] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_INFO] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_PITCH] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_SLICE] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_VIEW] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_ATTRIB] = true
	ContextRegisters[gcn.GREG_MM_CB_TARGET_MASK] = true

	ContextRegisters[gcn.GREG_MM_CB_BLEND0_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_CB_BLEND_RED] = true
	ContextRegisters[gcn.GREG_MM_CB_BLEND_GREEN] = true
	ContextRegisters[gcn.GREG_MM_CB_BLEND_BLUE] = true
	ContextRegisters[gcn.GREG_MM_CB_BLEND_ALPHA] = true

	ContextRegisters[gcn.GREG_MM_CB_COLOR0_CLEAR_WORD0] = true
	ContextRegisters[gcn.GREG_MM_CB_COLOR0_CLEAR_WORD1] = true

	ContextRegisters[gcn.GREG_MM_DB_RENDER_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_DB_SHADER_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_DB_DEPTH_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_DB_DEPTH_CLEAR] = true
	ContextRegisters[gcn.GREG_MM_DB_DEPTH_SIZE] = true
	ContextRegisters[gcn.GREG_MM_DB_Z_WRITE_BASE] = true
	ContextRegisters[gcn.GREG_MM_DB_Z_READ_BASE] = true
	ContextRegisters[gcn.GREG_MM_DB_Z_INFO] = true

	ContextRegisters[gcn.GREG_MM_DB_STENCIL_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_DB_STENCILREFMASK] = true
	ContextRegisters[gcn.GREG_MM_DB_STENCILREFMASK_BF] = true
	ContextRegisters[gcn.GREG_MM_DB_STENCIL_CLEAR] = true

	ContextRegisters[gcn.GREG_MM_SPI_PS_IN_CONTROL] = true
	ContextRegisters[gcn.GREG_MM_SPI_PS_INPUT_ENA] = true
	ContextRegisters[gcn.GREG_MM_SPI_PS_INPUT_ADDR] = true

	// Range covering SPI_PS_INPUT_CNTL_0 to SPI_PS_INPUT_CNTL_31
	for i := uint32(gcn.GREG_MM_SPI_PS_INPUT_CNTL_0); i <= gcn.GREG_MM_SPI_PS_INPUT_CNTL_31; i++ {
		ContextRegisters[i] = true
	}

	ContextRegisters[gcn.GREG_MM_SPI_SHADER_COL_FORMAT] = true
	ContextRegisters[gcn.GREG_MM_SPI_SHADER_Z_FORMAT] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VS_OUT_CNTL] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VTE_CNTL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_MODE_CNTL_0] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_VPORT_ZMIN_0] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_VPORT_ZMAX_0] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_XSCALE] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_XOFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_YSCALE] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_YOFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_ZSCALE] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_VPORT_ZOFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_CLIP_CNTL] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_GB_VERT_CLIP_ADJ] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_GB_HORZ_CLIP_ADJ] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_GB_VERT_DISC_ADJ] = true
	ContextRegisters[gcn.GREG_MM_PA_CL_GB_HORZ_DISC_ADJ] = true

	ContextRegisters[gcn.GREG_MM_PA_SC_SCREEN_SCISSOR_TL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_SCREEN_SCISSOR_BR] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_VPORT_SCISSOR_0_TL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_VPORT_SCISSOR_0_BR] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_GENERIC_SCISSOR_TL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_GENERIC_SCISSOR_BR] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_WINDOW_SCISSOR_TL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_WINDOW_SCISSOR_BR] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_WINDOW_OFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_SC_MODE_CNTL] = true

	ContextRegisters[gcn.GREG_MM_PA_SU_HARDWARE_SCREEN_OFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_LINE_STIPPLE_CNTL] = true

	ContextRegisters[gcn.GREG_MM_PA_SU_LINE_CNTL] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_AA_CONFIG] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_AA_MASK_X0Y0_X1Y0] = true
	ContextRegisters[gcn.GREG_MM_PA_SC_AA_MASK_X0Y1_X1Y1] = true

	ContextRegisters[gcn.GREG_MM_CB_SHADER_MASK] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_POLY_OFFSET_CLAMP] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_POLY_OFFSET_FRONT_SCALE] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_POLY_OFFSET_FRONT_OFFSET] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_POLY_OFFSET_BACK_SCALE] = true
	ContextRegisters[gcn.GREG_MM_PA_SU_POLY_OFFSET_BACK_OFFSET] = true

	ContextRegisters[gcn.GREG_MM_VGT_MULTI_PRIM_IB_RESET_INDX] = true
	ContextRegisters[gcn.GREG_MM_VGT_MULTI_PRIM_IB_RESET_EN] = true
}
