package np_manager

type NpState uint32

const (
	NpStateUnknown   = NpState(0)
	NpStateSignedOut = NpState(1)
	NpStateSignedIn  = NpState(2)
)

type NpOnlineId struct {
	Data  [16]byte
	Term  int8
	Dummy [3]byte
}

type NpId struct {
	Handle   NpOnlineId
	Opt      [8]byte
	Reserved [8]byte
}
