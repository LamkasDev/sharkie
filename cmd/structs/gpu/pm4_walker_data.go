package gpu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func (l *Liverpool) handleWriteData(ringName string, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] write data payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Check if we support the write destination.
	destSelection := (payload[0] >> 8) & 0x7
	switch destSelection {
	case 0, 1, 5:
	default:
		logger.Printf("[%s] failed write data on non-memory destination %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", destSelection),
		)
		return
	}

	// Get address of destination.
	addressLow := uint64(payload[1])
	addressHigh := uint64(payload[2] & 0xFFFF)
	address := uintptr(addressLow | (addressHigh << 32))
	if address == 0 {
		logger.Printf("[%s] failed write data invalid address.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Write data.
	data := payload[3:]
	dstSlice := unsafe.Slice((*uint32)(unsafe.Pointer(address)), len(data))
	copy(dstSlice, data)

	if LogPM4Packets {
		logger.Printf("[%s] wrote %s bytes to %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", len(data)),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
}

func (l *Liverpool) handleDmaData(ringName string, payload []uint32) {
	if len(payload) < 6 {
		logger.Printf("[%s] dma data payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Get address of source.
	srcAddrLow := uint64(payload[1])
	srcAddrHigh := uint64(payload[2])
	srcAddr := uintptr(srcAddrLow | (srcAddrHigh << 32))

	// Get address of destination.
	dstAddrLow := uint64(payload[3])
	dstAddrHigh := uint64(payload[4])
	dstAddr := uintptr(dstAddrLow | (dstAddrHigh << 32))

	// Validate.
	count := payload[5] & 0x3FFFFF
	if srcAddr == 0 || dstAddr == 0 {
		logger.Printf("[%s] failed dma data invalid address.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}

	// Copy the data.
	srcSlice := unsafe.Slice((*uint32)(unsafe.Pointer(srcAddr)), count)
	dstSlice := unsafe.Slice((*uint32)(unsafe.Pointer(dstAddr)), count)
	l.StateMutex.Lock()
	copy(dstSlice, srcSlice)
	l.StateMutex.Unlock()

	if LogPM4Packets {
		logger.Printf("[%s] copied %s bytes from %s to %s\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", count*4),
			color.Yellow.Sprintf("0x%X", srcAddr),
			color.Yellow.Sprintf("0x%X", dstAddr),
		)
	}
}

func (l *Liverpool) handleWriteConstRam(ringName string, payload []uint32) {
	if len(payload) < 1 {
		logger.Printf("[%s] write const ram payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
		)
		return
	}
	offset := int(payload[0] & 0xFFFF)
	data := payload[1:]
	if offset+len(data) > LiverpoolConstRamSize {
		logger.Printf("[%s] failed write const ram outside bounds (offset=%s, size=%s).\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Yellow.Sprintf("0x%X", offset),
			color.Green.Sprintf("%d", len(data)),
		)
		return
	}

	// Write data.
	l.StateMutex.Lock()
	copy(l.DrawState.ConstRam[offset:], data)
	l.StateMutex.Unlock()

	if true || LogPM4Packets {
		logger.Printf("[%s] wrote %s bytes to const ram at %s.\n",
			color.Green.Sprintf("PM4-%s/%d", ringName, len(payload)),
			color.Green.Sprintf("%d", len(data)),
			color.Yellow.Sprintf("0x%X", offset),
		)
	}
}
