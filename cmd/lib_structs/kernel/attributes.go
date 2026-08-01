package kernel

var GlobalPsfAttributes PsfAttributes

type PsfAttributes uint32

const (
	PsfAttributeSupportInitialUserLogout  = PsfAttributes(1 << 0)
	PsfAttributeEnterButtonCross          = PsfAttributes(1 << 1)
	PsfAttributePsMoveWarning             = PsfAttributes(1 << 2)
	PsfAttributeSupportStereoscopic3d     = PsfAttributes(1 << 3)
	PsfAttributePsButtonSuspend           = PsfAttributes(1 << 4)
	PsfAttributeEnterButtonSystem         = PsfAttributes(1 << 5)
	PsfAttributeOverrideShareMenu         = PsfAttributes(1 << 6)
	PsfAttributeSpecialResPsButtonSuspend = PsfAttributes(1 << 8)
	PsfAttributeEnableHdcp                = PsfAttributes(1 << 9)
	PsfAttributeDisableHdcpNonGame        = PsfAttributes(1 << 10)
	PsfAttributeSupportPsVr               = PsfAttributes(1 << 14)
	PsfAttributeSixCpuMode                = PsfAttributes(1 << 15)
	PsfAttributeSevenCpuMode              = PsfAttributes(1 << 16)
	PsfAttributeSupportNeoMode            = PsfAttributes(1 << 23)
	PsfAttributeRequirePsVr               = PsfAttributes(1 << 26)
	PsfAttributeSupportHdr                = PsfAttributes(1 << 29)
	PsfAttributeDisplayLocation           = PsfAttributes(1 << 31)
)

func (a PsfAttributes) Has(flag PsfAttributes) bool {
	return (a & flag) != 0
}

func (a *PsfAttributes) Set(flag PsfAttributes, enabled bool) {
	if enabled {
		*a |= flag
	} else {
		*a &^= flag
	}
}
