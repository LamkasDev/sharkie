package reg

type GpuMemoryBase uint32

func (r GpuMemoryBase) Address() uintptr {
	return (uintptr(r) << 8) & 0xFFFFFFFFFF
}
