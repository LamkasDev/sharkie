//go:generate go run ../gcn_gen/gcn_gen.go

package gcn

const (
	GcnRegBankSize = 0x400 // 1024 DWORDs per bank

	GcnRegBaseSystem     = 0x0000
	GcnRegBaseConfig     = 0x2000
	GcnRegBaseShader     = 0x2C00
	GcnRegBaseContext    = 0xA000
	GcnRegBaseUserConfig = 0xC000
)

const (
	// S_ENDPGM: type[31:23]=0b101111111 (SOPP), op[22:16]=1, simm16[15:0]=0
	GcnShaderEndProgramStandard = uint32(0xBF810000)

	// S_ENDPGM: raw encoding where bits[9:0] = 0x3FF
	GcnShaderEndProgramRaw = uint32(0xFF00FFFF)

	// S_SETPC_B64: sdst[22:16]=s0 ssrc0[7:0]=s0
	GcnShaderSetPcB64 = uint32(0xBE802000)

	// Maximum shader size we'll scan before giving up.
	GcnShaderMaxDwords = 16 * 1024
)

const (
	GcnDataFormatInvalid     = 0
	GcnDataFormat8           = 1
	GcnDataFormat16          = 2
	GcnDataFormat8_8         = 3
	GcnDataFormat32          = 4
	GcnDataFormat16_16       = 5
	GcnDataFormat10_11_11    = 6
	GcnDataFormat11_11_10    = 7
	GcnDataFormat10_10_10_2  = 8
	GcnDataFormat2_10_10_10  = 9
	GcnDataFormat8_8_8_8     = 10
	GcnDataFormat32_32       = 11
	GcnDataFormat16_16_16_16 = 12
	GcnDataFormat32_32_32    = 13
	GcnDataFormat32_32_32_32 = 14
	GcnDataFormatReserved_15 = 15
	GcnDataFormat5_6_5       = 16
	GcnDataFormat1_5_5_5     = 17
	GcnDataFormat5_5_5_1     = 18
	GcnDataFormat4_4_4_4     = 19
	GcnDataFormat8_24        = 20
	GcnDataFormat24_8        = 21
	GcnDataFormatX24_8_32    = 22

	GcnDataFormatBC1 = 35
	GcnDataFormatBC2 = 36
	GcnDataFormatBC3 = 37
	GcnDataFormatBC4 = 38
	GcnDataFormatBC5 = 39
	GcnDataFormatBC6 = 40
	GcnDataFormatBC7 = 41

	GcnDataFormatFmask8_1 = 47
)

const (
	GcnNumFormatUnorm     = 0
	GcnNumFormatSnorm     = 1
	GcnNumFormatUscaled   = 2
	GcnNumFormatSscaled   = 3
	GcnNumFormatUint      = 4
	GcnNumFormatSint      = 5
	GcnNumFormatSnormOgl  = 6
	GcnNumFormatSfloat    = 7
	GcnNumFormatReserved8 = 8
	GcnNumFormatSrgb      = 9
	GcnNumFormatUbnorm    = 10
	GcnNumFormatUbnormOgl = 11
	GcnNumFormatUbint     = 12
	GcnNumFormatUbscaled  = 13
	GcnNumFormatFloat     = 14
)

const (
	GcnImageTypeBuffer           = 0
	GcnImageTypeColor1D          = 8
	GcnImageTypeColor2D          = 9
	GcnImageTypeColor3D          = 10
	GcnImageTypeCubeOrArray      = 11
	GcnImageTypeColor1DArray     = 12
	GcnImageTypeColor2DArray     = 13
	GcnImageTypeColor2DMsaa      = 14
	GcnImageTypeColor2DMsaaArray = 15
)
