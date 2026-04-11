package gcn

const (
	SopkOpMovkI32        = 0x00
	SopkOpCmovkI32       = 0x02
	SopkOpCmpkEqI32      = 0x03
	SopkOpCmpkLgI32      = 0x04
	SopkOpCmpkGtI32      = 0x05
	SopkOpCmpkGeI32      = 0x06
	SopkOpCmpkLtI32      = 0x07
	SopkOpCmpkLeI32      = 0x08
	SopkOpCmpkEqU32      = 0x09
	SopkOpCmpkLgU32      = 0x0A
	SopkOpCmpkGtU32      = 0x0B
	SopkOpCmpkGeU32      = 0x0C
	SopkOpCmpkLtU32      = 0x0D
	SopkOpCmpkLeU32      = 0x0E
	SopkOpAddkI32        = 0x0F
	SopkOpMulkI32        = 0x10
	SopkOpCbranchIFork   = 0x11
	SopkOpGetregB32      = 0x12
	SopkOpSetregB32      = 0x13
	SopkOpSetregImm32B32 = 0x15
)
