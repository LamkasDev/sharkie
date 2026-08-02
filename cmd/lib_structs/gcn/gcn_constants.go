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

	// Maximum shader size we'll scan before giving up.
	GcnShaderMaxDwords = 16 * 1024
)
