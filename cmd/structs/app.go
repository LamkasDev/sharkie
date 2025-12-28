package structs

const CurrentFirmwareVersion = uint32(0x11008001)
const GameCompiledSdkVersion = uint32(0x04508001)

type AppInfo struct {
	AppId               int32
	MmapFlags           int32
	AttributeExecutable int32
	Attribute2          int32
	CusaName            [10]byte
	DebugLevel          uint8
	SlvFlags            uint8
	MiniAppDmemFlags    uint8
	RenderMode          uint8
	MdbgOut             uint8
	RequiredHdcpType    uint8
	PreloadPrxFlags     uint64
	Attribute1          int32
	HasParamSfo         int32
	TitleWorkaround     TitleWorkaround
}

type TitleWorkaround struct {
	Version int32
	Align   int32
	Ids     [2]uint64
}
