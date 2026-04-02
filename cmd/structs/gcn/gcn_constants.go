//go:generate go run ../gcn_gen/gcn_gen.go

package gcn

const (
	GcnRegBankSize = 0x400 // 1024 DWORDs per bank

	GcnRegBaseConfig     = 0x2000
	GcnRegBaseShader     = 0x2C00
	GcnRegBaseContext    = 0xA000
	GcnRegBaseUserConfig = 0xC000
)
