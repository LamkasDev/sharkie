package np_manager

type NpState uint32

const (
	NpStateUnknown   = NpState(0)
	NpStateSignedOut = NpState(1)
	NpStateSignedIn  = NpState(2)
)
