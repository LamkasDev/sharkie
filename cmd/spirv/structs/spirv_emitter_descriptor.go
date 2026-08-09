package structs

const (
	DescriptorSetSlotGlobal = 0
	DescriptorSetSlotImages = 1
)

const (
	GlobalBindingAddressTranslation = 0

	ImageBindingSampledImages1D      = 0
	ImageBindingStorageImages1D      = 1
	ImageBindingSampledImages2D      = 2
	ImageBindingStorageImages2D      = 3
	ImageBindingSampledImages3D      = 4
	ImageBindingStorageImages3D      = 5
	ImageBindingSampledImages2DArray = 6
	ImageBindingStorageImages2DArray = 7

	VertexBindingOffset = 64
	MaxStaticBindings   = 128
)
