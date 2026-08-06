package structs

const (
	DescriptorSetSlotStatic = 0
)

const (
	StaticBindingAddressTranslation   = 0
	StaticBindingSampledBuffers       = 1
	StaticBindingSampledImages1D      = 2
	StaticBindingStorageImages1D      = 3
	StaticBindingSampledImages2D      = 4
	StaticBindingStorageImages2D      = 5
	StaticBindingSampledImages3D      = 6
	StaticBindingStorageImages3D      = 7
	StaticBindingSampledImages2DArray = 8
	StaticBindingStorageImages2DArray = 9

	VertexBindingOffset = 64
	MaxStaticBindings   = 128
)
