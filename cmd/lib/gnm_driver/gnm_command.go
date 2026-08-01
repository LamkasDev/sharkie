package gnm_driver

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/video_out"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu/pm4"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000011B0
// __int64 __fastcall sceGnmSubmitCommandBuffers(__int64, __int64, __int64, __int64, __int64)
func libSceGnmDriver_sceGnmSubmitCommandBuffers(count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr *uintptr) int64 {
	return libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload(count, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr)
}

// 0x0000000000000F80
// __int64 __fastcall sceGnmSubmitCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64)
func libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload(workloadId, count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr *uintptr) int64 {
	if count == 0 {
		logger.Printf("%-132s %s skipped due to zero count.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		)
		return 0
	}
	if dcbGpuAddrsPtr == nil || dcbSizesPtr == nil || (ccbSizesPtr != nil && ccbGpuAddrsPtr == nil) {
		logger.Printf("%-132s %s failed due to invalid addresses pointer.\n",
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

	// Submit buffers.
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
		logger.Printf("%-132s %s submitted %s indirect buffers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
			color.Green.Sprintf("%d", len(buffers)),
		)
	}
	return 0
}

// 0x0000000000001690
// __int64 __fastcall sceGnmSubmitAndFlipCommandBuffers(__int64, __int64, __int64, __int64, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffers(count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr *uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) int64 {
	return libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload(count, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr, videoOutHandle, bufferIndex, flipMode, flipArg)
}

// 0x0000000000001410
// __int64 __fastcall sceGnmSubmitAndFlipCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload(workloadId, count uint32, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr *uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) int64 {
	if count == 0 {
		logger.Printf("%-132s %s skipped due to zero count.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return 0
	}
	if dcbGpuAddrsPtr == nil || dcbSizesPtr == nil || (ccbSizesPtr != nil && ccbGpuAddrsPtr == nil) {
		logger.Printf("%-132s %s failed due to invalid addresses pointer.\n",
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
	newDcbSizeDW, err := gnmPatchPrepareFlip(lastDcbAddress, lastDcbSizeDW, videoOutHandle, bufferIndex, flipMode, flipArg)
	if err != nil {
		logger.Printf("%-132s %s failed due to gnmPatchPrepareFlip error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_FLIP_FAILED
	}
	dcbSizes[lastIdx] = newDcbSizeDW * 4

	// Submit buffers.
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

	// Schedule the flip.
	flipResult := video_out.SceVideoOutSubmitEopFlip(uintptr(videoOutHandle), uintptr(bufferIndex), uintptr(flipMode), uintptr(flipArg), 0)
	if flipResult != 0 {
		logger.Printf("%-132s %s failed due to sceVideoOutSubmitEopFlip error.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_VALUE
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s submitted %s indirect buffers and requested flip.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
			color.Green.Sprintf("%d", len(buffers)),
		)
	}
	return 0
}

// 0x00000000000019A0
// __int64 __fastcall sceGnmRequestFlipAndSubmitDone(int, int, int, int, int, __int64)
func libSceGnmDriver_sceGnmRequestFlipAndSubmitDone(dcbPtr, requestId uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) uintptr {
	return libSceGnmDriver_sceGnmRequestFlipAndSubmitDoneForWorkload(dcbPtr, dcbPtr, requestId, videoOutHandle, bufferIndex, flipMode, flipArg)
}

// 0x00000000000017C0
// __int64 __fastcall sceGnmRequestFlipAndSubmitDoneForWorkload(__int64, __int64, unsigned int, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmRequestFlipAndSubmitDoneForWorkload(ctxPtr, dcbPtr, requestId uintptr, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) uintptr {
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

	// Write the minimal prepare flip header into the caller's buffer.
	nop := (*PM4CmdNop)(unsafe.Pointer(dcbPtr))
	nop.Header = GNM_PREPARE_FLIP_MAGIC
	nop.DataBlock[0] = GNM_PREPARE_FLIP_VARIANT_BASE

	// Patch prepare flip packet.
	newDcbSizeDW, err := gnmPatchPrepareFlip(dcbPtr, 64, videoOutHandle, bufferIndex, flipMode, flipArg)
	if err != nil {
		logger.Printf("%-132s %s failed due to gnmPatchPrepareFlip error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmRequestFlipAndSubmitDoneForWorkload"),
			err.Error(),
		)
		return SCE_GNM_ERROR_FLIP_FAILED
	}

	// Build a single IB packet pointing at the inline DCB.
	buffer := NewPM4IndirectBuffer(dcbPtr, newDcbSizeDW*4, false)
	buffers := []PM4IndirectBuffer{buffer}

	// Submit buffers and wait.
	GlobalLiverpool.SubmitCommandBuffers(buffers)
	GlobalLiverpool.WaitOnFence()

	// Signal that we're done.
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))

	// Schedule the flip.
	flipResult := video_out.SceVideoOutSubmitEopFlip(uintptr(videoOutHandle), uintptr(bufferIndex), uintptr(flipMode), uintptr(flipArg), 0)
	if flipResult != 0 {
		logger.Printf("%-132s %s failed due to sceVideoOutSubmitEopFlip error.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		)
		return SCE_GNM_ERROR_INVALID_VALUE
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s requested flip and signaled done.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmRequestFlipAndSubmitDoneForWorkload"),
		)
	}
	return 0
}

func gnmPatchPrepareFlip(lastDcbAddress uintptr, lastDcbSizeDW, videoOutHandle, bufferIndex, flipMode uint32, flipArg int64) (uint32, error) {
	if bufferIndex == 0xFFFFFFFF {
		return 0, fmt.Errorf("invalid buffer index")
	}
	if lastDcbSizeDW < GNM_PREPARE_FLIP_OFFSET_DWORDS {
		return 0, fmt.Errorf("last DCB too small to hold prepare flip block (%d DWORDs)", lastDcbSizeDW)
	}

	// The prepare flip packet starts 64 DWORDs before end of the last DCB.
	packetDWOffset := lastDcbSizeDW - GNM_PREPARE_FLIP_OFFSET_DWORDS
	packetPtr := lastDcbAddress + uintptr(packetDWOffset)*4
	packetBase := (*[GNM_PREPARE_FLIP_OFFSET_DWORDS]uint32)(unsafe.Pointer(packetPtr))
	if packetBase[0] != GNM_PREPARE_FLIP_MAGIC {
		return 0, fmt.Errorf("prepare flip header mismatch at DCB+%d (got 0x%X, want 0x%X)", packetDWOffset, packetBase[0], GNM_PREPARE_FLIP_MAGIC)
	}
	previous := make([]uint32, 7)
	copy(previous, packetBase[:7])
	variant := previous[1]

	// Get the handle's label buffer base address to build the WRITE_DATA target.
	var labelBase uintptr
	labelResult := video_out.SceVideoOutGetBufferLabelAddress(videoOutHandle, &labelBase)
	if labelResult != 0 || labelBase == 0 {
		logger.Printf(
			"%-132s %s skipping WRITE_DATA patch.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("gnmPatchPrepareFlip"),
		)
		return packetDWOffset, nil
	}
	labelAddress := labelBase + uintptr(bufferIndex)*8

	// Write data packet to signal surface.
	writeLock := (*PM4CmdWriteData)(unsafe.Pointer(packetPtr))
	writeLock.Header = NewPM4TypedHeader(PM4_IT_WRITE_DATA, 4)
	writeLock.Control = 0x500
	writeLock.AddressLow = uint32(labelAddress)
	writeLock.AddressHigh = uint32(labelAddress >> 32)
	writeLock.Data[0] = 1

	// NOP packet.
	nop := (*PM4CmdNop)(unsafe.Pointer(packetPtr + 5*4))

	switch variant {
	case 0x68750777: // PrepareFlip.
		nop.Header = NewPM4TypedHeader(PM4_IT_NOP, 0x3A)
		nop.DataBlock[0] = 0x68750776 // PatchedFlip.
	case 0x68750778: // PrepareFlipLabel.
		nop.Header = NewPM4TypedHeader(PM4_IT_NOP, 0x35)
		nop.DataBlock[0] = 0x68750776 // PatchedFlip.

		writeLabel := (*PM4CmdWriteData)(unsafe.Pointer(packetPtr + 0x3B*4))
		writeLabel.Header = NewPM4TypedHeader(PM4_IT_WRITE_DATA, 4)
		writeLabel.Control = 0x500
		writeLabel.AddressLow = previous[2] & 0xFFFFFFFC
		writeLabel.AddressHigh = previous[3]
		writeLabel.Data[0] = previous[4]
	case 0x68750780: // PrepareFlipInterrupt.
		nop.Header = NewPM4TypedHeader(PM4_IT_NOP, 0x34)
		nop.DataBlock[0] = 0x68750776 // PatchedFlip.

		writeEop := (*PM4CmdEventWriteEop)(unsafe.Pointer(packetPtr + 0x3A*4))
		writeEop.Header = NewPM4TypedHeader(PM4_IT_EVENT_WRITE_EOP, 5)
		writeEop.EventControl = (previous[2] & 0x3F) + 0x500 + (previous[3]&0x3F)*0x1000
		writeEop.AddressLow = 0
		writeEop.DataControl = 0x1000000
		writeEop.DataLow = 0
		writeEop.DataHigh = 0
	case 0x68750781: // PrepareFlipInterruptLabel.
		nop.Header = NewPM4TypedHeader(PM4_IT_NOP, 0x34)
		nop.DataBlock[0] = 0x68750776 // PatchedFlip.

		writeEop := (*PM4CmdEventWriteEop)(unsafe.Pointer(packetPtr + 0x3A*4))
		writeEop.Header = NewPM4TypedHeader(PM4_IT_EVENT_WRITE_EOP, 5)
		writeEop.EventControl = (previous[5] & 0x3F) + 0x500 + (previous[6]&0x3F)*0x1000
		writeEop.AddressLow = previous[2] & 0xFFFFFFFC
		writeEop.DataControl = (previous[3] & 0xFFFF) | 0x22000000
		writeEop.DataLow = previous[4]
		writeEop.DataHigh = 0
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s patched prepare flip variant %s at %s (label=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("gnmPatchPrepareFlip"),
			color.Yellow.Sprintf("0x%X", variant),
			color.Yellow.Sprintf("0x%X", packetPtr),
			color.Yellow.Sprintf("0x%X", labelAddress),
		)
	}
	return packetDWOffset + 64, nil
}

// 0x0000000000001720
// __int64 sceGnmSubmitDone()
func libSceGnmDriver_sceGnmSubmitDone() int64 {
	// Wait for work to finish.
	GlobalLiverpool.WaitOnFence()

	// Signal that we're done.
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))

	if logger.LogGraphics {
		logger.Printf("%-132s %s signaled done.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitDone"),
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

	if logger.LogGraphics {
		logger.Printf("%-132s %s dinged ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmDingDongForWorkload"),
			color.Green.Sprintf("%d", vqId),
		)
	}
	return 0
}
