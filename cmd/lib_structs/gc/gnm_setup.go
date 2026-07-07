package gc

import "unsafe"

const (
	SCE_GC_IOCTL_SET_RING_SIZES = 0xC00C8110
	SCE_GC_IOCTL_GET_CU_MASK    = 0xC010810B
)

type GnmComputeUnitMask struct {
	Mask1 uint32
	Mask2 uint32
	Mask3 uint32
	Mask4 uint32
}

const GnmComputeUnitMaskSize = unsafe.Sizeof(GnmComputeUnitMask{})

type GnmRingSizes struct {
	Ring1 uint32
	Ring2 uint32
	_     uint32
}

const GnmRingSizesSize = unsafe.Sizeof(GnmRingSizes{})
