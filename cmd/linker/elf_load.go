package linker

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

// ProcessLoadSection copies data the by PT_LOAD section into memory.
func ProcessLoadSection(e *elf.Elf, s *elf.ElfLoadSection, data []byte) {
	if s.PFilesz == 0 {
		return
	}

	if s.POffset+s.PFilesz > uint64(len(data)) {
		available := uint64(len(data)) - s.POffset
		copy(e.Memory[s.PVaddr:], data[s.POffset:s.POffset+available])
	} else {
		copy(e.Memory[s.PVaddr:], data[s.POffset:s.POffset+s.PFilesz])
	}
}
