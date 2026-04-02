package gpu

const (
	RegBankSize = 0x400 // 1024 DWORDs per bank

	RegBaseConfig     = uint32(0x2000)
	RegBaseShader     = uint32(0x2C00)
	RegBaseContext    = uint32(0xA000)
	RegBaseUserConfig = uint32(0xC000)
)

// LiverpoolRegisters mirrors register banks on the Liverpool GPU.
type LiverpoolRegisters struct {
	Config     [RegBankSize]uint32
	Shader     [RegBankSize]uint32
	Context    [RegBankSize]uint32
	UserConfig [RegBankSize]uint32
}

// LiverpoolDrawState holds state derived from SET_* packets that is needed to decode draw calls.
// It is reset when Walk() clears the rings.
type LiverpoolDrawState struct {
	IndexType uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase uintptr // CPU-visible address of the current index buffer
}
