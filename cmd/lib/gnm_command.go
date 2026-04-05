package lib

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gc"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	"github.com/gookit/color"
)

// 0x00000000000011B0
// __int64 __fastcall sceGnmSubmitCommandBuffers(__int64, __int64, __int64, __int64, __int64)
func libSceGnmDriver_sceGnmSubmitCommandBuffers(count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr) int64 {
	return libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload(count, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr)
}

// 0x0000000000000F80
// __int64 __fastcall sceGnmSubmitCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64)
func libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload(workloadId, count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr) int64 {
	if count == 0 {
		logger.Printf("%-132s %s skipped due to zero count.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		)
		return 0
	}
	if dcbGpuAddrsPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid DCB gpu addresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}
	if dcbSizesPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid DCB sizes pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}
	if ccbSizesPtr != 0 && ccbGpuAddrsPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid CCB gpu addresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}

	// Validate all DCB sizes.
	dcbSizes := unsafe.Slice((*uint32)(unsafe.Pointer(dcbSizesPtr)), count)
	for i := range count {
		dcbSize := dcbSizes[i]
		if dcbSize == 0 {
			logger.Printf("%-132s %s failed due to DCB %s having zero size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
				color.Yellow.Sprintf("%d", i),
			)
			return SCE_GNM_ERROR_INVALID_VALUE
		}
		if dcbSize>>2 > GNM_MAX_CB_SIZE_DWORDS {
			logger.Printf("%-132s %s failed due to DCB %s size exceeding limit.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
				color.Yellow.Sprintf("%d", i),
			)
			return SCE_GNM_ERROR_INVALID_VALUE
		}
	}

	// Rotate ring and submit buffers.
	GlobalGraphicsController.ActiveRingSlot++
	buffers, err := BuildPM4IndirectBuffers(count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr)
	if err != nil {
		logger.Printf("%-132s %s failed due to BuildPM4IndirectBuffers error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_INVALID_VALUE
	}
	GlobalLiverpool.SubmitCommandBuffers(buffers)

	if logger.LogGraphics {
		logger.Printf("%-132s %s submitted %s indirect buffers to ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
			color.Green.Sprintf("%d", len(buffers)),
			color.Green.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
	}
	return 0
}

// 0x0000000000001690
// __int64 __fastcall sceGnmSubmitAndFlipCommandBuffers(__int64, __int64, __int64, __int64, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffers(count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) int64 {
	return libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload(count, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr, videoOutHandle, bufferIndex, flipMode, flipArg)
}

// 0x0000000000001410
// __int64 __fastcall sceGnmSubmitAndFlipCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload(workloadId, count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) int64 {
	if count == 0 {
		logger.Printf("%-132s %s skipped due to zero count.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return 0
	}
	if dcbGpuAddrsPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid DCB gpu addresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}
	if dcbSizesPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid DCB sizes pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}
	if ccbSizesPtr != 0 && ccbGpuAddrsPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid CCB gpu addresses pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}

	// Validate all DCB sizes.
	dcbAddresses := unsafe.Slice((*uintptr)(unsafe.Pointer(dcbGpuAddrsPtr)), count)
	dcbSizes := unsafe.Slice((*uint32)(unsafe.Pointer(dcbSizesPtr)), count)
	for i := range count {
		dcbSize := dcbSizes[i]
		if dcbSize == 0 {
			logger.Printf("%-132s %s failed due to DCB %s having zero size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
				color.Yellow.Sprintf("%d", i),
			)
			return SCE_GNM_ERROR_INVALID_VALUE
		}
		if dcbSize>>2 > GNM_MAX_CB_SIZE_DWORDS {
			logger.Printf("%-132s %s failed due to DCB %s size exceeding limit.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
				color.Yellow.Sprintf("%d", i),
			)
			return SCE_GNM_ERROR_INVALID_VALUE
		}
	}

	// Patch prepare flip packet.
	lastIdx := count - 1
	lastDcbAddress := dcbAddresses[lastIdx]
	lastDcbSizeDW := dcbSizes[lastIdx] >> 2
	if err := gnmPatchPrepareFlip(lastDcbAddress, lastDcbSizeDW, videoOutHandle, bufferIndex, flipMode, flipArg); err != nil {
		logger.Printf("%-132s %s failed due to gnmPatchPrepareFlip error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_FLIP_FAILED
	}

	// Rotate ring and submit buffers.
	GlobalGraphicsController.ActiveRingSlot++
	buffers, err := BuildPM4IndirectBuffers(count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr)
	if err != nil {
		logger.Printf("%-132s %s failed due to BuildPM4IndirectBuffers error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_INVALID_VALUE
	}
	GlobalLiverpool.SubmitCommandBuffers(buffers)

	if logger.LogGraphics {
		logger.Printf("%-132s %s submitted %s indirect buffers to ring %s and requested flip.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
			color.Green.Sprintf("%d", len(buffers)),
			color.Green.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
	}
	return 0
}

// 0x00000000000019A0
// __int64 __fastcall sceGnmRequestFlipAndSubmitDone(int, int, int, int, int, __int64)
func libSceGnmDriver_sceGnmRequestFlipAndSubmitDone(dcbPtr, requestId, videoOutHandle, bufferIndex, flipMode, flipArg uintptr) uintptr {
	return libSceGnmDriver_sceGnmRequestFlipAndSubmitDoneForWorkload(dcbPtr, dcbPtr, requestId, videoOutHandle, bufferIndex, flipMode, flipArg)
}

// 0x00000000000017C0
// __int64 __fastcall sceGnmRequestFlipAndSubmitDoneForWorkload(__int64, __int64, unsigned int, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmRequestFlipAndSubmitDoneForWorkload(ctxPtr, dcbPtr, requestId, videoOutHandle, bufferIndex, flipMode, flipArg uintptr) uintptr {
	if requestId < 0x100 {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_VALUE
	}
	if dcbPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid DCB pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_POINTER
	}

	// Drain any queued ring work.
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_DRAIN_RING, 0)

	// Rotate ring slot.
	switchBuffer := GnmSwitchBuffer{
		RingSlot: GlobalGraphicsController.ActiveRingSlot + 1,
	}
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_SWITCH_BUFFER, uintptr(unsafe.Pointer(&switchBuffer)))

	// Write the minimal prepare flip header into the caller's buffer.
	pkt := (*[64]uint32)(unsafe.Pointer(dcbPtr))
	pkt[0] = GNM_PREPARE_FLIP_MAGIC
	pkt[1] = GNM_PREPARE_FLIP_VARIANT_BASE

	// Patch the prepare flip block and schedule it.
	if err := gnmPatchPrepareFlip(dcbPtr, uint32(len(pkt)), uint32(videoOutHandle), uint32(bufferIndex), uint32(flipMode), int64(flipArg)); err != nil {
		logger.Printf("%-132s %s failed due to gnmPatchPrepareFlip error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmRequestFlipAndSubmitDoneForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_FLIP_FAILED
	}

	// Build a single IB packet pointing at the inline DCB.
	buffer := NewPM4IndirectBuffer(dcbPtr, uint32(unsafe.Sizeof(pkt)), false)
	buffers := []PM4IndirectBuffer{buffer}

	// Submit it.
	GlobalLiverpool.SubmitCommandBuffers(buffers)

	// Flush the ring and mark it idle.
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_SUBMIT_DONE, 0)

	// Signal that we're done.
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))
	GlobalGraphicsController.RingActive = false
	GlobalGraphicsController.PendingSubmits = 0

	if logger.LogGraphics {
		logger.Printf("%-132s %s requested flip and signaled done on ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmRequestFlipAndSubmitDoneForWorkload"),
			color.Green.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
	}
	return 0
}

func gnmPatchPrepareFlip(lastDcbAddress uintptr, lastDcbSizeDW, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) error {
	if bufferIndex == 0xFFFFFFFF {
		return fmt.Errorf("invalid buffer index")
	}
	if lastDcbSizeDW < GNM_PREPARE_FLIP_OFFSET_DWORDS {
		return fmt.Errorf("last DCB too small to hold prepare flip block (%d DWORDs)", lastDcbSizeDW)
	}

	// The prepare flip packet starts 64 DWORDs before end of the last DCB.
	packetDWOffset := lastDcbSizeDW - GNM_PREPARE_FLIP_OFFSET_DWORDS
	packetPtr := lastDcbAddress + uintptr(packetDWOffset)*4
	packetBase := (*[GNM_PREPARE_FLIP_OFFSET_DWORDS]uint32)(unsafe.Pointer(packetPtr))
	if packetBase[0] != GNM_PREPARE_FLIP_MAGIC {
		return fmt.Errorf("prepare flip header mismatch at DCB+%d (got 0x%X, want 0x%X)", packetDWOffset, packetBase[0], GNM_PREPARE_FLIP_MAGIC)
	}
	variant := packetBase[1]
	if variant < GNM_PREPARE_FLIP_VARIANT_BASE || variant > GNM_PREPARE_FLIP_VARIANT_MAX {
		return fmt.Errorf("unknown prepare flip variant 0x%X", variant)
	}
	if variant == GNM_PREPARE_FLIP_VARIANT_ADDR && (packetBase[2]&3) != 0 {
		return fmt.Errorf("prepare flip variant ADDR gpu address 0x%X is not 4-byte aligned", packetBase[2])
	}

	// Schedule the flip.
	flipResult := libSceVideoOut_sceVideoOutSubmitEopFlip(uintptr(videoOutHandle), uintptr(bufferIndex), uintptr(flipMode), uintptr(flipArg), 0)
	if flipResult != 0 {
		return fmt.Errorf("sceVideoOutSubmitEopFlip returned 0x%X", flipResult)
	}

	// Get the handle's label buffer base address to build the WRITE_DATA target.
	var labelBase uintptr
	labelResult := libSceVideoOut_sceVideoOutGetBufferLabelAddress(uintptr(videoOutHandle), uintptr(unsafe.Pointer(&labelBase)))
	if labelResult != 0 || labelBase == 0 {
		// Label address unavailable - skip the WRITE_DATA patch.
		logger.Printf(
			"%-132s %s skipping WRITE_DATA patch.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("gnmPatchPrepareFlip"),
		)
		return nil
	}

	// Patch the prepare flip packet to a PM4 WRITE_DATA packet.
	labelAddress := labelBase + uintptr(bufferIndex)*8
	packetBase[0] = PM4_WRITE_DATA_HEADER
	packetBase[1] = PM4_WRITE_DATA_CONTROL
	packetBase[2] = uint32(labelAddress)
	packetBase[3] = uint32(labelAddress >> 32)
	packetBase[4] = 1

	if logger.LogGraphics {
		logger.Printf("%-132s %s patched prepare flip to WRITE_DATA at %s (label=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("gnmPatchPrepareFlip"),
			color.Yellow.Sprintf("0x%X", packetPtr),
			color.Yellow.Sprintf("0x%X", labelAddress),
		)
	}
	return nil
}

// 0x0000000000001720
// __int64 sceGnmSubmitDone()
func libSceGnmDriver_sceGnmSubmitDone() int64 {
	// Drain any queued ring work.
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_DRAIN_RING, 0)

	// Flush the ring and mark it idle.
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_SUBMIT_DONE, 0)

	// Signal that we're done.
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))
	GlobalGraphicsController.RingActive = false
	GlobalGraphicsController.PendingSubmits = 0

	if logger.LogGraphics {
		logger.Printf("%-132s %s signaled done on ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitDone"),
			color.Green.Sprintf("%d", GlobalGraphicsController.ActiveRingSlot),
		)
	}
	return 0
}

// TODO: this isn't right
// 0x0000000000004020
// __int64 __fastcall sceGnmDingDong(unsigned int a1, unsigned int a2)
func libSceGnmDriver_sceGnmDingDong(vqId, nextOffsetsDw uint32) int64 {
	return libSceGnmDriver_sceGnmDingDongForWorkload(vqId, nextOffsetsDw, 0)
}

// TODO: this isn't right
// 0x0000000000003F60
// __int64 __fastcall sceGnmDingDongForWorkload(unsigned int, unsigned int)
func libSceGnmDriver_sceGnmDingDongForWorkload(vqId, nextOffsetsDw uint32, workloadId uintptr) int64 {
	if vqId == 0 {
		logger.Printf("%-132s %s skipped due to invalid ring index.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmDingDongForWorkload"),
		)
		return 0
	}

	// Drain any queued ring work.
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_DRAIN_RING, 0)

	// Decode ring index into doorbell coordinates and issue write.
	ring := uint32(vqId) - 1
	dingDong := GnmDingDong{
		PipeIndex:    (ring >> 5) + 1,
		QueueIndex:   (ring & 0x1F) >> 3,
		SlotIndex:    ring & 0x07,
		WritePointer: nextOffsetsDw,
	}
	GlobalGraphicsController.Ioctl(SCE_GC_IOCTL_DINGDONG, uintptr(unsafe.Pointer(&dingDong)))

	if logger.LogGraphics {
		logger.Printf("%-132s %s dinged ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmDingDongForWorkload"),
			color.Green.Sprintf("%d", vqId),
		)
	}
	return 0
}
