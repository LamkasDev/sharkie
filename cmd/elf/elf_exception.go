package elf

import (
	"encoding/binary"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// ProcessExceptionFrameSection parses an existing exception frame section and stores information about the data.
func ProcessExceptionFrameSection(e *Elf) {
	// Now we need to actually parse the section and figure out the exception frame address.
	headerAddr := e.BaseAddress + uintptr(e.ExceptionFrameSection.PVaddr)
	memOffset := uintptr(e.ExceptionFrameSection.PVaddr)

	e.ExceptionFrameSection.Address = headerAddr
	e.ExceptionFrameSection.LoadedSize = e.ExceptionFrameSection.PMemsz

	// Ensure we can read the header.
	if uint64(memOffset+8) <= e.MemSize {
		encoding := e.Memory[memOffset+1]
		switch encoding {
		case EFRAME_PCREL | EFRAME_SDATA4:
			relOffset := int32(binary.LittleEndian.Uint32(e.Memory[memOffset+4:]))
			dataAddr := uintptr(int64(headerAddr) + 4 + int64(relOffset))
			e.ExceptionFrameDataAddress = dataAddr

			// Not sure how big it is, really. Let's just let it run until 0.
			if dataAddr >= e.BaseAddress {
				offset := uint64(dataAddr - e.BaseAddress)
				if offset < e.MemSize {
					e.ExceptionFrameDataSize = e.MemSize - offset
				}
			}

			logger.Printf("Resolved %s data via header (headerAddr=%s, dataAddr=%s, size=%s).\n",
				color.Blue.Sprint(".eh_frame"),
				color.Yellow.Sprintf("0x%X", headerAddr),
				color.Yellow.Sprintf("0x%X", dataAddr),
				color.Green.Sprint(e.ExceptionFrameDataSize),
			)
		default:
			logger.Print(color.Gray.Sprintf(
				"Unknown .eh_frame_hdr encoding 0x%X, assuming data follows header.\n",
				encoding,
			))
			e.ExceptionFrameDataAddress = headerAddr + uintptr(e.ExceptionFrameSection.PMemsz)
		}
	}
}
