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

// 0x0000000000000F80
// __int64 __fastcall sceGnmSubmitCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64)
func libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload(workloadId, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr) uintptr {
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
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))

	logger.Printf("%-132s %s submitted %s indirect buffers (count=%s, workloadId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceGnmSubmitCommandBuffersForWorkload"),
		color.Green.Sprintf("%d", len(buffers)),
		color.Green.Sprintf("%d", count),
		color.Yellow.Sprintf("0x%X", workloadId),
	)
	return 0
}

// 0x0000000000001410
// __int64 __fastcall sceGnmSubmitAndFlipCommandBuffersForWorkload(__int64, __int64, __int64, __int64, __int64, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload(workloadId, count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr, videoOutHandle, bufferIndex, flipMode, flipArg uintptr) uintptr {
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
	if err := gnmPatchPrepareFlip(lastDcbAddress, lastDcbSizeDW, uint32(videoOutHandle), uint32(bufferIndex), uint32(flipMode), uint64(flipArg)); err != nil {
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
	WriteAddress(GlobalGraphicsController.SubmitDoneAddress, uintptr(1))

	logger.Printf("%-132s %s submitted %s indirect buffers (count=%s, workloadId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceGnmSubmitAndFlipCommandBuffersForWorkload"),
		color.Green.Sprintf("%d", len(buffers)),
		color.Green.Sprintf("%d", count),
		color.Yellow.Sprintf("0x%X", workloadId),
	)
	return 0
}

func gnmPatchPrepareFlip(lastDcbAddress uintptr, lastDcbSizeDW, videoOutHandle, bufferIndex, flipMode uint32, flipArg uint64) error {
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

	logger.Printf("%-132s %s patched prepare flip to WRITE_DATA at %s (label=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("gnmPatchPrepareFlip"),
		color.Yellow.Sprintf("0x%X", packetPtr),
		color.Yellow.Sprintf("0x%X", labelAddress),
	)
	return nil
}
