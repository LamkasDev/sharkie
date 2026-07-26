package structs

import (
	"slices"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
)

type VMA struct {
	Start      uintptr
	End        uintptr
	Prot       uint32
	Mapped     bool
	Name       string
	IsDirect   bool
	Offset     uint64
	MemoryType int32
}

func (m *MemoryManager) splitVMA(addr uintptr) {
	if addr == 0 || addr == ^uintptr(0) {
		return
	}
	for i := 0; i < len(m.VMAs); i++ {
		if m.VMAs[i].Start < addr && m.VMAs[i].End > addr {
			v1 := m.VMAs[i]
			v1.End = addr
			v2 := m.VMAs[i]
			v2.Start = addr
			m.VMAs = slices.Replace(m.VMAs, i, i+1, v1, v2)
			return
		}
	}
}

func (m *MemoryManager) updateVMA(start, end uintptr, update func(*VMA)) {
	start = start &^ uintptr(posix.SystemPageSize-1)
	end = (end + uintptr(posix.SystemPageSize-1)) &^ uintptr(posix.SystemPageSize-1)

	m.splitVMA(start)
	m.splitVMA(end)
	for i := 0; i < len(m.VMAs); i++ {
		if m.VMAs[i].Start >= end {
			break
		}
		if m.VMAs[i].End > start {
			update(&m.VMAs[i])
		}
	}

	if len(m.VMAs) == 0 {
		return
	}
	merged := []VMA{m.VMAs[0]}
	for i := 1; i < len(m.VMAs); i++ {
		last := &merged[len(merged)-1]
		v := m.VMAs[i]
		if last.End == v.Start && last.Prot == v.Prot && last.Mapped == v.Mapped && last.Name == v.Name && last.MemoryType == v.MemoryType {
			last.End = v.End
		} else {
			merged = append(merged, v)
		}
	}
	m.VMAs = merged
}

func (m *MemoryManager) VirtualQuery(addr uintptr, flags int32, info *posix.VirtualQueryInfo) uintptr {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for i := 0; i < len(m.VMAs); i++ {
		vma := m.VMAs[i]
		if addr >= vma.Start && addr < vma.End {
			if !vma.Mapped {
				if flags == 1 {
					for j := i + 1; j < len(m.VMAs); j++ {
						if m.VMAs[j].Mapped {
							vma = m.VMAs[j]
							goto found
						}
					}
				}
				return 0x8002000D // ORBIS_KERNEL_ERROR_EACCES
			}
		found:
			info.Start = uint64(vma.Start)
			info.End = uint64(vma.End)
			info.Offset = vma.Offset
			if !vma.Mapped {
				info.Protection = 0
			} else {
				info.Protection = int32(vma.Prot)
			}
			info.MemoryType = vma.MemoryType
			info.Bitfield = 0
			if vma.IsDirect {
				info.Bitfield |= 2 // is_direct
			}
			if vma.Mapped {
				info.Bitfield |= 16 // is_committed
			}
			copy(info.Name[:], []byte(vma.Name))
			return 0
		}
	}
	return 0x8002000D
}

func (m *MemoryManager) splitDirectVMA(addr uintptr) {
	if addr == 0 || addr == ^uintptr(0) {
		return
	}
	for i := 0; i < len(m.DirectVMAs); i++ {
		if m.DirectVMAs[i].Start < addr && m.DirectVMAs[i].End > addr {
			v1 := m.DirectVMAs[i]
			v1.End = addr
			v2 := m.DirectVMAs[i]
			v2.Start = addr
			m.DirectVMAs = slices.Replace(m.DirectVMAs, i, i+1, v1, v2)
			return
		}
	}
}

func (m *MemoryManager) updateDirectVMA(start, end uintptr, update func(*VMA)) {
	m.splitDirectVMA(start)
	m.splitDirectVMA(end)
	for i := 0; i < len(m.DirectVMAs); i++ {
		if m.DirectVMAs[i].Start >= end {
			break
		}
		if m.DirectVMAs[i].End > start {
			update(&m.DirectVMAs[i])
		}
	}

	if len(m.DirectVMAs) == 0 {
		return
	}
	merged := []VMA{m.DirectVMAs[0]}
	for i := 1; i < len(m.DirectVMAs); i++ {
		last := &merged[len(merged)-1]
		v := m.DirectVMAs[i]
		if last.End == v.Start && last.Prot == v.Prot && last.Mapped == v.Mapped && last.Name == v.Name && last.MemoryType == v.MemoryType {
			last.End = v.End
		} else {
			merged = append(merged, v)
		}
	}
	m.DirectVMAs = merged
}

func (m *MemoryManager) AllocateDirect(offset uintptr, length uintptr, memType int32) {
	end := offset + length
	m.Lock.Lock()
	m.updateDirectVMA(offset, end, func(v *VMA) {
		v.Mapped = true
		v.Prot = posix.PROT_READ | posix.PROT_WRITE
		v.IsDirect = true
		v.MemoryType = memType
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) DirectMemoryQuery(addr uintptr, flags int32, info *posix.VirtualQueryInfo) uintptr {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for i := 0; i < len(m.DirectVMAs); i++ {
		vma := m.DirectVMAs[i]
		if addr >= vma.Start && addr < vma.End {
			if !vma.Mapped {
				if flags == 1 {
					for j := i + 1; j < len(m.DirectVMAs); j++ {
						if m.DirectVMAs[j].Mapped {
							vma = m.DirectVMAs[j]
							goto found
						}
					}
				}
				return 0x8002000D // ORBIS_KERNEL_ERROR_EACCES
			}
		found:
			info.Start = uint64(vma.Start)
			info.End = uint64(vma.End)
			info.Offset = vma.Offset
			if !vma.Mapped {
				info.Protection = 0
			} else {
				info.Protection = int32(vma.Prot)
			}
			info.MemoryType = vma.MemoryType
			info.Bitfield = 0
			if vma.IsDirect {
				info.Bitfield |= 2 // is_direct
			}
			if vma.Mapped {
				info.Bitfield |= 16 // is_committed
			}
			copy(info.Name[:], []byte(vma.Name))
			return 0
		}
	}
	return 0x8002000D
}
