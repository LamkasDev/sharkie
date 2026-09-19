package structs

import "unsafe"

type PushConstants struct {
	UserDataAddress         uint64
	OnionMemoryBaseAddress  uint64
	GarlicMemoryBaseAddress uint64
	GdsMemoryBaseAddress    uint64

	UserSgprCount uint32
	ShaderRsrc2   uint32
	VteControl    uint32
	ClipControl   uint32

	VpXScale  float32
	VpXOffset float32
	VpYScale  float32
	VpYOffset float32
}

const PushConstantsSize = uint32(unsafe.Sizeof(PushConstants{}))

func init() {
	// Vertex and fragment each get a copy; 128 bytes is the Vulkan spec minimum.
	if PushConstantsSize*2 > 128 {
		panic("PushConstants exceed the 128-byte Vulkan minimum")
	}
}

const (
	MaxCommandsPerFrame             = 6144
	MaxSampleImageBindingsPerFrame  = MaxCommandsPerFrame * 8
	MaxStorageImageBindingsPerFrame = MaxCommandsPerFrame * 4
)

const (
	UserDataBufferSize = MaxCommandsPerFrame * UserDataSize
)

const (
	AddressTranslationBlockEntrySize = 8
	AddressTranslationBlockEntries   = MaxStaticBindings
	AddressTranslationBlockSize      = AddressTranslationBlockEntries * AddressTranslationBlockEntrySize
	AddressTranslationBufferSize     = AddressTranslationBlockSize * MaxCommandsPerFrame
)
