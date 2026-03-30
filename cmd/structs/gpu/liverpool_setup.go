package gpu

import "unsafe"

type LiverpoolCommandRing struct {
	Pending []PM4IndirectBuffer
}

const LiverpoolCommandRingSize = unsafe.Sizeof(LiverpoolCommandRing{})
