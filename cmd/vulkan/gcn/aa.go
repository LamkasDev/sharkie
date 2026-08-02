package gcn

import vk "github.com/goki/vulkan"

func TranslateMsaaSamples(msaaSampleLocations uint32) vk.SampleCountFlagBits {
	switch msaaSampleLocations {
	case 0:
		return vk.SampleCount1Bit
	case 1:
		return vk.SampleCount2Bit
	case 2:
		return vk.SampleCount4Bit
	case 3:
		return vk.SampleCount8Bit
	case 4:
		return vk.SampleCount16Bit
	default:
		return vk.SampleCount1Bit
	}
}
