package structs

const (
	DescriptorSetSlotBindless       = 0
	DescriptorSetSlotDiscovery      = 1
	DescriptorSetSlotTexel          = 2
	DescriptorSetSlotTexelSecondary = 3

	// Bindings for DescriptorSetSlotBindless (Set 0)
	BindlessBindingSampledImages = 0
	BindlessBindingStorageImages = 1

	// Bindings for DescriptorSetSlotDiscovery (Set 1)
	DiscoveryBindingMap             = 0
	DiscoveryBindingMissingResource = 1
)
