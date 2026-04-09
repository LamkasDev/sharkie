package gcn

import (
	"fmt"
	"strings"

	"go101.org/nstd"
)

var Mnemotics = map[Encoding]map[uint32]string{
	EncSOP2: {
		0x00: "s_add_u32",
		0x01: "s_sub_u32",
		0x02: "s_add_i32",
		0x03: "s_sub_i32",
		0x04: "s_addc_u32",
		0x05: "s_subb_u32",
		0x06: "s_min_i32",
		0x07: "s_min_u32",
		0x08: "s_max_i32",
		0x09: "s_max_u32",
		0x0A: "s_cselect_b32",
		0x0B: "s_cselect_b64",
		0x0E: "s_and_b32",
		0x0F: "s_and_b64",
		0x10: "s_or_b32",
		0x11: "s_or_b64",
		0x12: "s_xor_b32",
		0x13: "s_xor_b64",
		0x14: "s_andn2_b32",
		0x15: "s_andn2_b64",
		0x16: "s_orn2_b32",
		0x17: "s_orn2_b64",
		0x18: "s_nand_b32",
		0x19: "s_nand_b64",
		0x1A: "s_nor_b32",
		0x1B: "s_nor_b64",
		0x1C: "s_xnor_b32",
		0x1D: "s_xnor_b64",
		0x1E: "s_lshl_b32",
		0x1F: "s_lshl_b64",
		0x20: "s_lshr_b32",
		0x21: "s_lshr_b64",
		0x22: "s_ashr_i32",
		0x23: "s_ashr_i64",
		0x24: "s_bfm_b32",
		0x25: "s_bfm_b64",
		0x26: "s_mul_i32",
		0x27: "s_bfe_u32",
		0x28: "s_bfe_i32",
		0x29: "s_bfe_u64",
		0x2A: "s_bfe_i64",
		0x2B: "s_cbranch_g_fork",
		0x2C: "s_absdiff_i32",
	},
	EncSOPK: {
		0x00: "s_movk_i32",
		0x02: "s_cmovk_i32",
		0x03: "s_cmpk_eq_i32",
		0x04: "s_cmpk_lg_i32",
		0x05: "s_cmpk_gt_i32",
		0x06: "s_cmpk_ge_i32",
		0x07: "s_cmpk_lt_i32",
		0x08: "s_cmpk_le_i32",
		0x09: "s_cmpk_eq_u32",
		0x0A: "s_cmpk_lg_u32",
		0x0B: "s_cmpk_gt_u32",
		0x0C: "s_cmpk_ge_u32",
		0x0D: "s_cmpk_lt_u32",
		0x0E: "s_cmpk_le_u32",
		0x0F: "s_addk_i32",
		0x10: "s_mulk_i32",
		0x11: "s_cbranch_i_fork",
		0x12: "s_getreg_b32",
		0x13: "s_setreg_b32",
		0x15: "s_setreg_imm32_b32",
	},
	EncSOP1: {
		0x03: "s_mov_b32",
		0x04: "s_mov_b64",
		0x05: "s_cmov_b32",
		0x06: "s_cmov_b64",
		0x07: "s_not_b32",
		0x08: "s_not_b64",
		0x09: "s_wqm_b32",
		0x0A: "s_wqm_b64",
		0x0B: "s_brev_b32",
		0x0C: "s_brev_b64",
		0x0D: "s_bcnt0_i32_b32",
		0x0E: "s_bcnt0_i32_b64",
		0x0F: "s_bcnt1_i32_b32",
		0x10: "s_bcnt1_i32_b64",
		0x11: "s_ff0_i32_b32",
		0x12: "s_ff0_i32_b64",
		0x13: "s_ff1_i32_b32",
		0x14: "s_ff1_i32_b64",
		0x15: "s_flbit_i32_b32",
		0x16: "s_flbit_i32_b64",
		0x18: "s_ff0_i32_b64",
		0x1A: "s_ff1_i32_b64",
		0x1B: "s_bitset0_b32",
		0x1C: "s_bitset0_b64",
		0x1D: "s_bitset1_b32",
		0x1E: "s_bitset1_b64",
		0x1F: "s_getpc_b64",
		0x20: "s_setpc_b64",
		0x21: "s_swappc_b64",
		0x22: "s_rfe_b64",
		0x23: "s_flbit_i32",
		0x24: "s_flbit_i32_i64",
		0x25: "s_flbit_i32_b32",
		0x26: "s_flbit_i32_b64",
		0x27: "s_and_saveexec_b64",
		0x28: "s_or_saveexec_b64",
		0x29: "s_sext_i32_i8",
		0x2A: "s_sext_i32_i16",
		0x2B: "s_xnor_saveexec_b64",
		0x2C: "s_quadmask_b32",
		0x2D: "s_quadmask_b64",
		0x2E: "s_movrels_b32",
		0x2F: "s_movrels_b64",
		0x30: "s_movreld_b32",
		0x31: "s_movreld_b64",
		0x32: "s_cbranch_join",
		0x34: "s_abs_i32",
		0x36: "s_and_saveexec_b64",
		0x37: "s_or_saveexec_b64",
		0x38: "s_xor_saveexec_b64",
		0x39: "s_andn2_saveexec_b64",
		0x40: "s_orn2_saveexec_b64",
		0x41: "s_nand_saveexec_b64",
		0x42: "s_nor_saveexec_b64",
		0x43: "s_xnor_saveexec_b64",
	},
	EncSOPC: {
		0x00: "s_cmp_eq_i32",
		0x01: "s_cmp_lg_i32",
		0x02: "s_cmp_gt_i32",
		0x03: "s_cmp_ge_i32",
		0x04: "s_cmp_lt_i32",
		0x05: "s_cmp_le_i32",
		0x06: "s_cmp_eq_u32",
		0x07: "s_cmp_lg_u32",
		0x08: "s_cmp_gt_u32",
		0x09: "s_cmp_ge_u32",
		0x0A: "s_cmp_lt_u32",
		0x0B: "s_cmp_le_u32",
		0x0C: "s_bitcmp0_b32",
		0x0D: "s_bitcmp1_b32",
		0x0E: "s_bitcmp0_b64",
		0x0F: "s_bitcmp1_b64",
		0x10: "s_setvskip",
	},
	EncSOPP: {
		0x00: "s_nop",
		0x01: "s_endpgm",
		0x02: "s_branch",
		0x04: "s_cbranch_scc0",
		0x05: "s_cbranch_scc1",
		0x06: "s_cbranch_vccz",
		0x07: "s_cbranch_vccnz",
		0x08: "s_cbranch_execz",
		0x09: "s_cbranch_execnz",
		0x0A: "s_barrier",
		0x0B: "s_setkill",
		0x0C: "s_waitcnt",
		0x0D: "s_sethalt",
		0x0E: "s_sleep",
		0x0F: "s_setprio",
		0x10: "s_sendmsg",
		0x11: "s_sendmsghalt",
		0x12: "s_trap",
		0x13: "s_icache_inv",
		0x14: "s_incperflevel",
		0x15: "s_decperflevel",
		0x16: "s_ttracedata",
		0x17: "s_cbranch_cdbgsys",
		0x18: "s_cbranch_cdbguser",
		0x19: "s_cbranch_cdbgsys_or_user",
		0x1A: "s_cbranch_cdbgsys_and_user",
	},
	EncVOP2: {
		0x00: "v_cndmask_b32",
		0x01: "v_readlane_b32",
		0x02: "v_writelane_b32",
		0x03: "v_add_f32",
		0x04: "v_sub_f32",
		0x05: "v_subrev_f32",
		0x06: "v_mac_legacy_f32",
		0x07: "v_mul_legacy_f32",
		0x08: "v_mul_f32",
		0x09: "v_mul_i32_i24",
		0x0A: "v_mul_hi_i32_i24",
		0x0B: "v_mul_u32_u24",
		0x0C: "v_mul_hi_u32_u24",
		0x0D: "v_min_legacy_f32",
		0x0E: "v_max_legacy_f32",
		0x0F: "v_min_f32",
		0x10: "v_max_f32",
		0x11: "v_min_i32",
		0x12: "v_max_i32",
		0x13: "v_min_u32",
		0x14: "v_max_u32",
		0x15: "v_lshr_b32",
		0x16: "v_lshrrev_b32",
		0x17: "v_ashr_i32",
		0x18: "v_ashrrev_i32",
		0x19: "v_lshl_b32",
		0x1A: "v_lshlrev_b32",
		0x1B: "v_and_b32",
		0x1C: "v_or_b32",
		0x1D: "v_xor_b32",
		0x1E: "v_bfm_b32",
		0x1F: "v_mac_f32",
		0x20: "v_madmk_f32",
		0x21: "v_madak_f32",
		0x22: "v_bcnt_u32_b32",
		0x23: "v_mbcnt_lo_u32_b32",
		0x24: "v_mbcnt_hi_u32_b32",
		0x25: "v_add_i32",
		0x26: "v_sub_i32",
		0x27: "v_subrev_i32",
		0x28: "v_addc_u32",
		0x29: "v_subb_u32",
		0x2A: "v_subbrev_u32",
		0x2B: "v_ldexp_f32",
		0x2C: "v_cvt_pkaccum_u8_f32",
		0x2D: "v_cvt_pknorm_i16_f32",
		0x2E: "v_cvt_pknorm_u16_f32",
		0x2F: "v_cvt_pkrtz_f16_f32",
		0x30: "v_cvt_pk_u16_u32",
		0x31: "v_cvt_pk_i16_i32",
	},
	EncVOP1: {
		0x00: "v_nop",
		0x01: "v_mov_b32",
		0x02: "v_readfirstlane_b32",
		0x03: "v_cvt_i32_f64",
		0x04: "v_cvt_f64_i32",
		0x05: "v_cvt_f32_i32",
		0x06: "v_cvt_f32_u32",
		0x07: "v_cvt_u32_f32",
		0x08: "v_cvt_i32_f32",
		0x0A: "v_cvt_f16_f32",
		0x0B: "v_cvt_f32_f16",
		0x0C: "v_cvt_rpi_i32_f32",
		0x0D: "v_cvt_flr_i32_f32",
		0x0E: "v_cvt_off_f32_i4",
		0x0F: "v_cvt_f32_f64",
		0x10: "v_cvt_f64_f32",
		0x11: "v_cvt_f32_ubyte0",
		0x12: "v_cvt_f32_ubyte1",
		0x13: "v_cvt_f32_ubyte2",
		0x14: "v_cvt_f32_ubyte3",
		0x15: "v_cvt_u32_f64",
		0x16: "v_cvt_f64_u32",
		0x17: "v_trunc_f64",
		0x18: "v_ceil_f64",
		0x19: "v_rndne_f64",
		0x1A: "v_floor_f64",
		0x20: "v_fract_f32",
		0x21: "v_trunc_f32",
		0x22: "v_ceil_f32",
		0x23: "v_rndne_f32",
		0x24: "v_floor_f32",
		0x25: "v_exp_f32",
		0x26: "v_log_clamp_f32",
		0x27: "v_log_f32",
		0x2A: "v_rcp_f32",
		0x2B: "v_rcp_clamp_f32",
		0x2D: "v_rsq_legacy_f32",
		0x2E: "v_rsq_f32",
		0x2F: "v_rcp_f64",
		0x30: "v_rcp_clamp_f64",
		0x31: "v_rsq_f64",
		0x32: "v_rsq_clamp_f64",
		0x33: "v_sqrt_f32",
		0x34: "v_sqrt_f64",
		0x35: "v_sin_f32",
		0x36: "v_cos_f32",
		0x37: "v_not_b32",
		0x38: "v_bfrev_b32",
		0x39: "v_ffbh_u32",
		0x3A: "v_ffbl_b32",
		0x3B: "v_ffbh_i32",
		0x3C: "v_frexp_exp_i32_f64",
		0x3D: "v_frexp_mant_f64",
		0x3E: "v_fract_f64",
		0x3F: "v_frexp_exp_i32_f32",
		0x40: "v_frexp_mant_f32",
		0x41: "v_clrexcp",
		0x42: "v_movreld_b32",
		0x43: "v_movrels_b32",
		0x44: "v_movrelsd_b32",
		0x45: "v_log_legacy_f32",
		0x46: "v_exp_legacy_f32",
	},
	EncVOPC: VopcMap(),
	EncVINTRP: {
		0x2: "v_interp_mov_f32",
		0x0: "v_interp_p1_f32",
		0x1: "v_interp_p2_f32",
	},
	EncVOP3: {
		// VOP2 opcodes.
		0x100: "v_cndmask_b32",
		0x101: "v_readlane_b32",
		0x102: "v_writelane_b32",
		0x103: "v_add_f32",
		0x104: "v_sub_f32",
		0x105: "v_subrev_f32",
		0x106: "v_mac_legacy_f32",
		0x107: "v_mul_legacy_f32",
		0x108: "v_mul_f32",
		0x109: "v_mul_i32_i24",
		0x10A: "v_mul_hi_i32_i24",
		0x10B: "v_mul_u32_u24",
		0x10C: "v_mul_hi_u32_u24",
		0x10D: "v_min_legacy_f32",
		0x10E: "v_max_legacy_f32",
		0x10F: "v_min_f32",
		0x110: "v_max_f32",
		0x111: "v_min_i32",
		0x112: "v_max_i32",
		0x113: "v_min_u32",
		0x114: "v_max_u32",
		0x115: "v_lshr_b32",
		0x116: "v_lshrrev_b32",
		0x117: "v_ashr_i32",
		0x118: "v_ashrrev_i32",
		0x119: "v_lshl_b32",
		0x11A: "v_lshlrev_b32",
		0x11B: "v_and_b32",
		0x11C: "v_or_b32",
		0x11D: "v_xor_b32",
		0x11E: "v_bfm_b32",
		0x11F: "v_mac_f32",
		0x122: "v_bcnt_u32_b32",
		0x124: "v_mbcnt_hi_u32_b32",
		0x125: "v_add_i32",
		0x126: "v_sub_i32",
		0x127: "v_subrev_i32",
		0x128: "v_addc_u32",
		0x129: "v_subb_u32",
		0x12A: "v_subbrev_u32",
		0x12B: "v_ldexp_f32",
		0x12C: "v_cvt_pkaccum_u8_f32",
		0x12D: "v_cvt_pknorm_i16_f32",
		0x12E: "v_cvt_pknorm_u16_f32",
		0x12F: "v_cvt_pkrtz_f16_f32",
		0x130: "v_cvt_pk_u16_u32",
		0x131: "v_cvt_pk_i16_i32",

		// VOP3 opcodes.
		0x140: "v_mad_legacy_f32",
		0x141: "v_mad_f32",
		0x142: "v_mad_i32_i24",
		0x143: "v_mad_u32_u24",
		0x144: "v_cubeid_f32",
		0x145: "v_cubesc_f32",
		0x146: "v_cubetc_f32",
		0x147: "v_cubema_f32",
		0x148: "v_bfe_u32",
		0x149: "v_bfe_i32",
		0x14A: "v_bfi_b32",
		0x14B: "v_fma_f32",
		0x14C: "v_fma_f64",
		0x14D: "v_lerp_u8",
		0x14E: "v_alignbit_b32",
		0x14F: "v_alignbyte_b32",
		0x150: "v_mullit_f32",
		0x151: "v_min3_f32",
		0x152: "v_min3_i32",
		0x153: "v_min3_u32",
		0x154: "v_max3_f32",
		0x155: "v_max3_i32",
		0x156: "v_max3_u32",
		0x157: "v_med3_f32",
		0x158: "v_med3_i32",
		0x159: "v_med3_u32",
		0x15A: "v_sad_u8",
		0x15B: "v_sad_hi_u8",
		0x15C: "v_sad_u16",
		0x15D: "v_sad_u32",
		0x15E: "v_cvt_pk_u8_f32",
		0x15F: "v_div_fixup_f32",
		0x160: "v_div_fixup_f64",
		0x161: "v_lshl_b64",
		0x162: "v_lshr_b64",
		0x163: "v_ashr_i64",
		0x164: "v_add_f64",
		0x165: "v_mul_f64",
		0x166: "v_min3_f64",
		0x167: "v_max_f64",
		0x168: "v_ldexp_f64",
		0x169: "v_mul_lo_u32",
		0x16A: "v_mul_hi_u32",
		0x16B: "v_mul_lo_i32",
		0x16C: "v_mul_hi_i32",
		0x16D: "v_div_scale_f32",
		0x16E: "v_div_scale_f64",
		0x16F: "v_div_fmas_f32",
		0x170: "v_div_fmas_f64",
		0x171: "v_msad_u8",
		0x172: "v_qsad_pk_u16_u8",
		0x173: "v_mqsad_pk_u16_u8",
		0x174: "v_trig_preop_f64",
		0x175: "v_mqsad_u32_u8",
		0x176: "v_mad_u64_u32",
		0x177: "v_mad_i64_i32",

		// VOP1 opcodes.
		0x180: "v_nop",
		0x181: "v_mov_b32",
		0x182: "v_readfirstlane_b32",
		0x183: "v_cvt_i32_f64",
		0x184: "v_cvt_f64_i32",
		0x185: "v_cvt_f32_i32",
		0x186: "v_cvt_f32_u32",
		0x187: "v_cvt_u32_f32",
		0x188: "v_cvt_i32_f32",
		0x18A: "v_cvt_f16_f32",
		0x18B: "v_cvt_f32_f16",
		0x18C: "v_cvt_rpi_i32_f32",
		0x18D: "v_cvt_flr_i32_f32",
		0x18E: "v_cvt_off_f32_i4",
		0x18F: "v_cvt_f32_f64",
		0x190: "v_cvt_f64_f32",
		0x191: "v_cvt_f32_ubyte0",
		0x192: "v_cvt_f32_ubyte1",
		0x193: "v_cvt_f32_ubyte2",
		0x194: "v_cvt_f32_ubyte3",
		0x195: "v_cvt_u32_f64",
		0x196: "v_cvt_f64_u32",
		0x1A0: "v_fract_f32",
		0x1A1: "v_trunc_f32",
		0x1A2: "v_ceil_f32",
		0x1A3: "v_rndne_f32",
		0x1A4: "v_floor_f32",
		0x1A5: "v_exp_f32",
		0x1A6: "v_log_clamp_f32",
		0x1A7: "v_log_f32",
		0x1A8: "v_rcp_clamp_f32",
		0x1A9: "v_rcp_legacy_f32",
		0x1AA: "v_rcp_f32",
		0x1AB: "v_rcp_iflag_f32",
		0x1AC: "v_rsq_clamp_f32",
		0x1AD: "v_rsq_legacy_f32",
		0x1AE: "v_rsq_f32",
		0x1AF: "v_rcp_f64",
		0x1B0: "v_rcp_clamp_f64",
		0x1B1: "v_rsq_f64",
		0x1B2: "v_rsq_clamp_f64",
		0x1B3: "v_sqrt_f32",
		0x1B4: "v_sqrt_f64",
		0x1B5: "v_sin_f32",
		0x1B6: "v_cos_f32",
		0x1B7: "v_not_b32",
		0x1B8: "v_bfrev_b32",
		0x1B9: "v_ffbh_u32",
		0x1BA: "v_ffbl_b32",
		0x1BB: "v_ffbh_i32",
		0x1BC: "v_frexp_exp_i32_f64",
		0x1BD: "v_frexp_mant_f64",
		0x1BE: "v_fract_f64",
		0x1BF: "v_frexp_exp_i32_f32",
		0x1C0: "v_frexp_mant_f32",
		0x1C1: "v_clrexcp",
		0x1C2: "v_movreld_b32",
		0x1C4: "v_movrelsd_b32",
	},
	EncSMRD: {
		0x00: "s_load_dword",
		0x01: "s_load_dwordx2",
		0x02: "s_load_dwordx4",
		0x03: "s_load_dwordx8",
		0x04: "s_load_dwordx16",
		0x08: "s_buffer_load_dword",
		0x09: "s_buffer_load_dwordx2",
		0x0A: "s_buffer_load_dwordx4",
		0x0B: "s_buffer_load_dwordx8",
		0x0C: "s_buffer_load_dwordx16",
		0x1D: "s_dcache_inv_vol",
		0x1E: "s_memtime",
		0x1F: "s_dcache_inv",
	},
	EncMUBUF: {
		0x00: "buffer_load_format_x",
		0x01: "buffer_load_format_xy",
		0x02: "buffer_load_format_xyz",
		0x03: "buffer_load_format_xyzw",
		0x04: "buffer_store_format_x",
		0x05: "buffer_store_format_xy",
		0x06: "buffer_store_format_xyz",
		0x07: "buffer_store_format_xyzw",
		0x08: "buffer_load_ubyte",
		0x09: "buffer_load_sbyte",
		0x0A: "buffer_load_ushort",
		0x0B: "buffer_load_sshort",
		0x0C: "buffer_load_dword",
		0x0D: "buffer_load_dwordx2",
		0x0E: "buffer_load_dwordx4",
		0x0F: "buffer_load_dwordx3",
		0x18: "buffer_store_byte",
		0x1A: "buffer_store_short",
		0x1C: "buffer_store_dword",
		0x1D: "buffer_store_dwordx2",
		0x1E: "buffer_store_dwordx4",
		0x1F: "buffer_store_dwordx3",
		0x30: "buffer_atomic_swap",
		0x31: "buffer_atomic_cmpswap",
		0x32: "buffer_atomic_add",
		0x33: "buffer_atomic_sub",
		0x35: "buffer_atomic_smin",
		0x36: "buffer_atomic_umin",
		0x37: "buffer_atomic_smax",
		0x38: "buffer_atomic_umax",
		0x39: "buffer_atomic_and",
		0x3A: "buffer_atomic_or",
		0x3B: "buffer_atomic_xor",
		0x3C: "buffer_atomic_inc",
		0x3D: "buffer_atomic_dec",
		0x3E: "buffer_atomic_fcmpswap",
		0x3F: "buffer_atomic_fmin",
		0x40: "buffer_atomic_fmax",
		0x50: "buffer_atomic_swap_x2",
		0x51: "buffer_atomic_cmpswap_x2",
		0x52: "buffer_atomic_add_x2",
		0x53: "buffer_atomic_sub_x2",
		0x55: "buffer_atomic_smin_x2",
		0x56: "buffer_atomic_umin_x2",
		0x57: "buffer_atomic_smax_x2",
		0x58: "buffer_atomic_umax_x2",
		0x59: "buffer_atomic_and_x2",
		0x5A: "buffer_atomic_or_x2",
		0x5B: "buffer_atomic_xor_x2",
		0x5C: "buffer_atomic_inc_x2",
		0x5D: "buffer_atomic_dec_x2",
		0x5E: "buffer_atomic_fcmpswap_x2",
		0x5F: "buffer_atomic_fmin_x2",
		0x60: "buffer_atomic_fmax_x2",
		0x70: "buffer_wbinvl1_vol",
		0x71: "buffer_wbinvl1",
	},
	EncDS: {
		0x00: "ds_add_u32",
		0x01: "ds_sub_u32",
		0x02: "ds_rsub_u32",
		0x03: "ds_inc_u32",
		0x04: "ds_dec_u32",
		0x05: "ds_min_i32",
		0x06: "ds_max_i32",
		0x07: "ds_min_u32",
		0x08: "ds_max_u32",
		0x09: "ds_and_b32",
		0x0A: "ds_or_b32",
		0x0B: "ds_xor_b32",
		0x0C: "ds_mskor_b32",
		0x0D: "ds_write_b32",
		0x0E: "ds_write2_b32",
		0x0F: "ds_write2st64_b32",
		0x10: "ds_cmpst_b32",
		0x11: "ds_cmpst_f32",
		0x12: "ds_min_f32",
		0x13: "ds_max_f32",
		0x14: "ds_nop",
		0x18: "ds_gws_sema_release_all",
		0x19: "ds_gws_init",
		0x1A: "ds_gws_sema_v",
		0x1B: "ds_gws_sema_br",
		0x1C: "ds_gws_sema_p",
		0x1D: "ds_gws_barrier",
		0x1E: "ds_write_b8",
		0x1F: "ds_write_b16",
		0x20: "ds_add_rtn_u32",
		0x21: "ds_sub_rtn_u32",
		0x22: "ds_rsub_rtn_u32",
		0x23: "ds_inc_rtn_u32",
		0x24: "ds_dec_rtn_u32",
		0x25: "ds_min_rtn_i32",
		0x26: "ds_max_rtn_i32",
		0x27: "ds_min_rtn_u32",
		0x28: "ds_max_rtn_u32",
		0x29: "ds_and_rtn_b32",
		0x2A: "ds_or_rtn_b32",
		0x2B: "ds_xor_rtn_b32",
		0x2C: "ds_mskor_rtn_b32",
		0x2D: "ds_wrxchg_rtn_b32",
		0x2E: "ds_wrxchg2_rtn_b32",
		0x2F: "ds_wrxchg2st64_rtn_b32",
		0x30: "ds_cmpst_rtn_b32",
		0x31: "ds_cmpst_rtn_f32",
		0x32: "ds_min_rtn_f32",
		0x33: "ds_max_rtn_f32",
		0x34: "ds_wrap_rtn_b32",
		0x35: "ds_swizzle_b32",
		0x36: "ds_read_b32",
		0x37: "ds_read2_b32",
		0x38: "ds_read2st64_b32",
		0x39: "ds_read_i8",
		0x3A: "ds_read_u8",
		0x3B: "ds_read_i16",
		0x3C: "ds_read_u16",
		0x3D: "ds_consume",
		0x3E: "ds_append",
		0x3F: "ds_ordered_count",
		0x40: "ds_add_u64",
		0x41: "ds_sub_u64",
		0x42: "ds_rsub_u64",
		0x43: "ds_inc_u64",
		0x44: "ds_dec_u64",
		0x45: "ds_min_i64",
		0x46: "ds_max_i64",
		0x47: "ds_min_u64",
		0x48: "ds_max_u64",
		0x49: "ds_and_b64",
		0x4A: "ds_or_b64",
		0x4B: "ds_xor_b64",
		0x4C: "ds_mskor_b64",
		0x4D: "ds_write_b64",
		0x4E: "ds_write2_b64",
		0x4F: "ds_write2st64_b64",
		0x50: "ds_cmpst_b64",
		0x51: "ds_cmpst_f64",
		0x52: "ds_min_f64",
		0x53: "ds_max_f64",
		0x60: "ds_add_rtn_u64",
		0x61: "ds_sub_rtn_u64",
		0x62: "ds_rsub_rtn_u64",
		0x63: "ds_inc_rtn_u64",
		0x64: "ds_dec_rtn_u64",
		0x65: "ds_min_rtn_i64",
		0x66: "ds_max_rtn_i64",
		0x67: "ds_min_rtn_u64",
		0x68: "ds_max_rtn_u64",
		0x69: "ds_and_rtn_b64",
		0x6A: "ds_or_rtn_b64",
		0x6B: "ds_xor_rtn_b64",
		0x6C: "ds_mskor_rtn_b64",
		0x6D: "ds_wrxchg_rtn_b64",
		0x6E: "ds_wrxchg2_rtn_b64",
		0x6F: "ds_wrxchg2st64_rtn_b64",
		0x70: "ds_cmpst_rtn_b64",
		0x71: "ds_cmpst_rtn_f64",
		0x72: "ds_min_rtn_f64",
		0x73: "ds_max_rtn_f64",
		0x76: "ds_read_b64",
		0x77: "ds_read2_b64",
		0x78: "ds_read2st64_b64",
		0x7E: "ds_condxchg32_rtn_b64",
		0x80: "ds_add_src2_u32",
		0x81: "ds_sub_src2_u32",
		0x82: "ds_rsub_src2_u32",
		0x83: "ds_inc_src2_u32",
		0x84: "ds_dec_src2_u32",
		0x85: "ds_min_src2_i32",
		0x86: "ds_max_src2_i32",
		0x87: "ds_min_src2_u32",
		0x88: "ds_max_src2_u32",
		0x89: "ds_and_src2_b32",
		0x8A: "ds_or_src2_b32",
		0x8B: "ds_xor_src2_b32",
		0x8C: "ds_write_src2_b32",
		0x92: "ds_min_src2_f32",
		0x93: "ds_max_src2_f32",
		0xC0: "ds_add_src2_u64",
		0xC1: "ds_sub_src2_u64",
		0xC2: "ds_rsub_src2_u64",
		0xC3: "ds_inc_src2_u64",
		0xC4: "ds_dec_src2_u64",
		0xC5: "ds_min_src2_i64",
		0xC6: "ds_max_src2_i64",
		0xC7: "ds_min_src2_u64",
		0xC8: "ds_max_src2_u64",
		0xC9: "ds_and_src2_b64",
		0xCA: "ds_or_src2_b64",
		0xCB: "ds_xor_src2_b64",
		0xCC: "ds_write_src2_b64",
		0xD2: "ds_min_src2_f64",
		0xD3: "ds_max_src2_f64",
		0xDE: "ds_write_b96",
		0xDF: "ds_write_b128",
		0xFD: "ds_condxchg32_rtn_b128",
		0xFE: "ds_read_b96",
		0xFF: "ds_read_b128",
	},
	EncMIMG: {
		0x00: "image_load",
		0x01: "image_load_mip",
		0x02: "image_load_pck",
		0x03: "image_load_pck_sgn",
		0x04: "image_load_mip_pck",
		0x05: "image_load_mip_pck_sgn",
		0x08: "image_store",
		0x09: "image_store_mip",
		0x0A: "image_store_pck",
		0x0B: "image_store_mip_pck",
		0x0E: "image_get_resinfo",
		0x0F: "image_atomic_swap",
		0x10: "image_atomic_cmpswap",
		0x11: "image_atomic_add",
		0x12: "image_atomic_sub",
		0x14: "image_atomic_smin",
		0x15: "image_atomic_umin",
		0x16: "image_atomic_smax",
		0x17: "image_atomic_umax",
		0x18: "image_atomic_and",
		0x19: "image_atomic_or",
		0x1A: "image_atomic_xor",
		0x1B: "image_atomic_inc",
		0x1C: "image_atomic_dec",
		0x1D: "image_atomic_fcmpswap",
		0x1E: "image_atomic_fmin",
		0x1F: "image_atomic_fmax",
		0x20: "image_sample",
		0x21: "image_sample_cl",
		0x22: "image_sample_d",
		0x23: "image_sample_d_cl",
		0x24: "image_sample_l",
		0x25: "image_sample_b",
		0x26: "image_sample_b_cl",
		0x27: "image_sample_lz",
		0x28: "image_sample_c",
		0x29: "image_sample_c_cl",
		0x2A: "image_sample_c_d",
		0x2B: "image_sample_c_d_cl",
		0x2C: "image_sample_c_l",
		0x2D: "image_sample_c_b",
		0x2E: "image_sample_c_b_cl",
		0x2F: "image_sample_c_lz",
		0x30: "image_sample_o",
		0x31: "image_sample_cl_o",
		0x32: "image_sample_d_o",
		0x33: "image_sample_d_cl_o",
		0x34: "image_sample_l_o",
		0x35: "image_sample_b_o",
		0x36: "image_sample_b_cl_o",
		0x37: "image_sample_lz_o",
		0x38: "image_sample_c_o",
		0x39: "image_sample_c_cl_o",
		0x3A: "image_sample_c_d_o",
		0x3B: "image_sample_c_d_cl_o",
		0x3C: "image_sample_c_l_o",
		0x3D: "image_sample_c_b_o",
		0x3E: "image_sample_c_b_cl_o",
		0x3F: "image_sample_c_lz_o",
		0x40: "image_gather4",
		0x41: "image_gather4_cl",
		0x42: "image_gather4_l",
		0x43: "image_gather4_b",
		0x44: "image_gather4_b_cl",
		0x45: "image_gather4_lz",
		0x46: "image_gather4_c",
		0x47: "image_gather4_c_cl",
		0x4C: "image_gather4_c_l",
		0x4D: "image_gather4_c_b",
		0x4E: "image_gather4_c_b_cl",
		0x4F: "image_gather4_c_lz",
		0x50: "image_gather4_o",
		0x51: "image_gather4_cl_o",
		0x54: "image_gather4_l_o",
		0x55: "image_gather4_b_o",
		0x56: "image_gather4_b_cl_o",
		0x57: "image_gather4_lz_o",
		0x58: "image_gather4_c_o",
		0x59: "image_gather4_c_cl_o",
		0x5C: "image_gather4_c_l_o",
		0x5D: "image_gather4_c_b_o",
		0x5E: "image_gather4_c_b_cl_o",
		0x5F: "image_gather4_c_lz_o",
		0x60: "image_get_lod",
		0x68: "image_sample_cd",
		0x69: "image_sample_cd_cl",
		0x6A: "image_sample_c_cd",
		0x6B: "image_sample_c_cd_cl",
		0x6C: "image_sample_cd_o",
		0x6D: "image_sample_cd_cl_o",
		0x6E: "image_sample_c_cd_o",
		0x6F: "image_sample_c_cd_cl_o",
	},
	EncEXP: {},
}

func (instr *Instruction) GetMnemotic() string {
	var b string
	switch instr.Encoding {
	case EncSOP2, EncSOPK, EncSOP1, EncSOPC, EncSOPP:
		b = Mnemotics[instr.Encoding][instr.SOp]
	case EncVOP2, EncVOP1, EncVOPC:
		b = Mnemotics[instr.Encoding][instr.VOp]
	case EncVOP3:
		if instr.VOp < 0x100 {
			b = Mnemotics[EncVOPC][instr.VOp]
		} else {
			b = Mnemotics[instr.Encoding][instr.VOp]
		}
	case EncVINTRP:
	case EncSMRD:
		b = Mnemotics[instr.Encoding][instr.SmOp]
	case EncMTBUF, EncMUBUF:
		b = Mnemotics[instr.Encoding][instr.VmbOp]
	case EncMIMG:
		b = Mnemotics[instr.Encoding][instr.VmiOp]
	case EncDS:
	case EncEXP:
		b = "export"
	}
	if len(b) == 0 {
		return "??"
	}

	return b
}

func (instr *Instruction) GetFieldsString() string {
	var b strings.Builder
	switch instr.Encoding {
	case EncSOP2:
		fmt.Fprintf(&b, "sdst=%s ssrc1=%s ssrc0=%s",
			OperandToString(instr.SDst),
			OperandToString(instr.SSrc1),
			OperandToString(instr.SSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncSOPK:
		fmt.Fprintf(&b, "sdst=%s simm=0x%04X",
			OperandToString(instr.SDst),
			instr.Simm16,
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncSOP1:
		fmt.Fprintf(&b, "sdst=%s ssrc0=%s",
			OperandToString(instr.SDst),
			OperandToString(instr.SSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncSOPC:
		fmt.Fprintf(&b, "ssrc1=%s ssrc0=%s",
			OperandToString(instr.SSrc1),
			OperandToString(instr.SSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncSOPP:
		fmt.Fprintf(&b, "simm=0x%04X",
			instr.Simm16,
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncVOP2:
		fmt.Fprintf(&b, "vdst=%d vsrc1=%d src0=%s",
			instr.VDst,
			instr.VSrc1,
			OperandToString(instr.VSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncVOP1:
		fmt.Fprintf(&b, "vdst=%d src0=%s",
			instr.VDst,
			OperandToString(instr.VSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncVOPC:
		fmt.Fprintf(&b, "vsrc1=%d src0=%s",
			instr.VSrc1,
			OperandToString(instr.VSrc0),
		)
		if instr.HasLiteral {
			fmt.Fprintf(&b, " lit=0x%08X", instr.Literal)
		}
	case EncVINTRP:
	case EncVOP3:
		fmt.Fprintf(&b, "neg=%d omod=%d src2=%s src1=%s src0=%s clamp=%d abs=%d sdst=%s vdst=%s",
			instr.VNeg,
			instr.VOMod,
			OperandToString(instr.VSrc2),
			OperandToString(instr.VSrc1),
			OperandToString(instr.VSrc0),
			nstd.Btoi(instr.VClamp),
			instr.VAbs,
			OperandToString(instr.VSdst),
			OperandToString(instr.VDst),
		)
	case EncSMRD:
		fmt.Fprintf(&b, "sdst=%d sbase=%d imm=%d offset=%d",
			instr.SmDst,
			instr.SmBase,
			nstd.Btoi(instr.SmImmOff),
			instr.SmOffset,
		)
	case EncMTBUF:
	case EncMUBUF:
	case EncMIMG:
		fmt.Fprintf(&b, "ssamp=%d srsrc=%d vdata=%d vaddr=%d slc=%d lwe=%d tfe=%d r128=%d da=%d glc=%d unrm=%d dmask=%d",
			instr.VmiSsamp,
			instr.VmiSrsrc,
			instr.VmiVdata,
			instr.VmiVaddr,
			nstd.Btoi(instr.VmiSlc),
			nstd.Btoi(instr.VmiLwe),
			nstd.Btoi(instr.VmiTfe),
			nstd.Btoi(instr.VmiR128),
			nstd.Btoi(instr.VmiDa),
			nstd.Btoi(instr.VmiGlc),
			nstd.Btoi(instr.VmiUnorm),
			instr.VmiDmask,
		)
	case EncDS:
	case EncEXP:
		fmt.Fprintf(&b, "vsrc3=%d vsrc2=%d vsrc1=%d vsrc0=%d vm=%d done=%d compr=%d target=%d en=%d",
			instr.ExpVSrcs[3],
			instr.ExpVSrcs[2],
			instr.ExpVSrcs[1],
			instr.ExpVSrcs[0],
			nstd.Btoi(instr.ExpVm),
			nstd.Btoi(instr.ExpDone),
			nstd.Btoi(instr.ExpCompr),
			instr.ExpTarget,
			instr.ExpEn,
		)
	}
	if b.Len() == 0 {
		return "??"
	}

	return b.String()
}

func VopcMap() map[uint32]string {
	m := make(map[uint32]string)
	add16 := func(base uint32, prefix, suffix string) {
		ops := []string{
			"f", "lt", "eq", "le", "gt", "lg", "ge", "o",
			"u", "nge", "nlg", "ngt", "nle", "neq", "nlt", "tru",
		}
		for i, op := range ops {
			opcode := base + uint32(i)
			m[opcode] = "v_cmp" + prefix + "_" + op + suffix
		}
	}
	add8 := func(base uint32, prefix, suffix string) {
		ops := []string{"f", "lt", "eq", "le", "gt", "lg", "ge", "tru"}
		for i, op := range ops {
			opcode := base + uint32(i)
			m[opcode] = "v_cmp" + prefix + "_" + op + suffix
		}
	}

	// VOPC Instructions with 16 Compare Operations.
	add16(0x00, "", "_f32")  // V_CMP_{OP16}_F32
	add16(0x10, "x", "_f32") // V_CMPX_{OP16}_F32
	add16(0x20, "", "_f64")  // V_CMP_{OP16}_F64
	add16(0x30, "x", "_f64") // V_CMPX_{OP16}_F64

	add16(0x40, "s", "_f32")  // V_CMPS_{OP16}_F32
	add16(0x50, "sx", "_f32") // V_CMPSX_{OP16}_F32
	add16(0x60, "s", "_f64")  // V_CMPS_{OP16}_F64
	add16(0x70, "sx", "_f64") // V_CMPSX_{OP16}_F64

	// VOPC Instructions with Eight Compare Operations.
	add8(0x80, "", "_i32")  // V_CMP_{OP8}_I32
	add8(0x90, "x", "_i32") // V_CMPX_{OP8}_I32
	add8(0xA0, "", "_i64")  // V_CMP_{OP8}_I64
	add8(0xB0, "x", "_i64") // V_CMPX_{OP8}_I64

	add8(0xC0, "", "_u32")  // V_CMP_{OP8}_U32
	add8(0xD0, "x", "_u32") // V_CMPX_{OP8}_U32
	add8(0xE0, "", "_u64")  // V_CMP_{OP8}_U64
	add8(0xF0, "x", "_u64") // V_CMPX_{OP8}_U64

	m[0x88] = "v_cmp_class_f32"
	m[0x98] = "v_cmpx_class_f32"
	m[0xA8] = "v_cmp_class_f64"
	m[0xB8] = "v_cmpx_class_f64"

	return m
}

// Based on 5.2 Scalar ALU Operands.
func OperandToString(op uint32) string {
	switch {
	case op <= 103:
		return fmt.Sprintf("s%d", op)
	case op == 104:
		return "flat_scratch_lo"
	case op == 105:
		return "flat_scratch_hi"
	case op == 106:
		return "vcc_lo"
	case op == 107:
		return "vcc_hi"
	case op == 108:
		return "tba_lo"
	case op == 109:
		return "tba_hi"
	case op == 110:
		return "tma_lo"
	case op == 111:
		return "tma_hi"
	case op >= 112 && op <= 123:
		return fmt.Sprintf("ttmp%d", op-112)
	case op == 124:
		return "m0"
	case op == 126:
		return "exec_lo"
	case op == 127:
		return "exec_hi"
	case op == 128:
		return "0"
	case op >= 129 && op <= 192:
		return fmt.Sprint(op - 128)
	case op >= 193 && op <= 208:
		return fmt.Sprintf("-%d", op-192)
	case op == 240:
		return "0.5"
	case op == 241:
		return "-0.5"
	case op == 242:
		return "1.0"
	case op == 243:
		return "-1.0"
	case op == 244:
		return "2.0"
	case op == 245:
		return "-2.0"
	case op == 246:
		return "4.0"
	case op == 247:
		return "-4.0"
	case op == 251:
		return "vccz"
	case op == 252:
		return "execz"
	case op == 253:
		return "scc"
	case op == 255:
		return "lit"
	case op >= 256 && op <= 511:
		return fmt.Sprintf("v%d", op-256)
	}

	return fmt.Sprintf("?? (%d)", op)
}
