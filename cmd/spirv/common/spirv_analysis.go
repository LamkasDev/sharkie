package common

type SgprSource struct {
	UserDataOffset int32 // -1 if unknown
}

type SpirvShaderResource struct {
	InstructionOffset uintptr
	FixedSlot         int32 // -1 if dynamic
	RsrcUserData      int32 // UserData offset for T# base
	SampUserData      int32 // UserData offset for S# base
}
