package structs

type AuthInfo struct {
	ProcessId       uint64
	Capabilities    [4]uint64
	Attributes      [4]uint64
	UserCredentials [8]uint64
}
