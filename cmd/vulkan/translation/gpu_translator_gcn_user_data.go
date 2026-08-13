package translation

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/gookit/color"
)

func (t *GpuTranslator) UpdateUserDataBuffers(stream *gpu.LiverpoolCommandStream) {
	t.userDataBuffersMutex.Lock()
	defer t.userDataBuffersMutex.Unlock()

	// Find unique hashes in current draw calls or dispatches.
	activeHashes := make([]uint32, 0)
	hashesSeen := make(map[uint32]bool)
	for i := range stream.Draws {
		hash := stream.Draws[i].UserDataHash
		if !hashesSeen[hash] {
			activeHashes = append(activeHashes, hash)
			hashesSeen[hash] = true
		}
	}
	for i := range stream.Dispatches {
		hash := stream.Dispatches[i].UserDataHash
		if !hashesSeen[hash] {
			activeHashes = append(activeHashes, hash)
			hashesSeen[hash] = true
		}
	}

	// Reset offsets and map the user data buffer.
	clear(t.userDataOffsets)
	offset := uint32(0)

	// Create buffers for new active hashes.
	for _, hash := range activeHashes {
		// Get contents from global state.
		contents, ok := gpu.GlobalUserDataSnapshots[hash]
		if !ok {
			continue
		}

		// Upload.
		size := uint32(len(contents) * 4)
		if offset+size > uint32(structs.UserDataBufferSize) {
			logger.Printf("[%s] User data buffer overflow!\n", color.Red.Sprint("GPU"))
			break
		}

		copy(t.userDataBufferData[offset:], unsafe.Slice((*byte)(unsafe.Pointer(&contents[0])), size))
		t.userDataOffsets[hash] = offset
		offset += size

		/* logger.Printf("[%s] Created user data with hash %s (vtx=%x, frag=%x, cmp=%x).\n",
			color.Blue.Sprint("GPU"),
			color.Yellow.Sprintf("0x%X", hash),
			contents[gpu.UserDataOffsetVertex:gpu.UserDataOffsetHull],
			contents[gpu.UserDataOffsetFragment:gpu.UserDataOffsetCompute],
			contents[gpu.UserDataOffsetCompute:gpu.UserDataOffsetCompute+16],
		) */
	}
}
