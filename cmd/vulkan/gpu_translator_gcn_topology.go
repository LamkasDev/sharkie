package vulkan

import vk "github.com/goki/vulkan"

func translateTopology(primType uint32) vk.PrimitiveTopology {
	switch primType {
	case 1: // POINTLIST
		return vk.PrimitiveTopologyPointList
	case 2: // LINELIST
		return vk.PrimitiveTopologyLineList
	case 3: // LINESTRIP
		return vk.PrimitiveTopologyLineStrip
	case 4: // TRILIST
		return vk.PrimitiveTopologyTriangleList
	case 5: // TRIFAN
		return vk.PrimitiveTopologyTriangleFan
	case 6: // TRISTRIP
		return vk.PrimitiveTopologyTriangleStrip
	case 17: // RECTLIST
		return vk.PrimitiveTopologyTriangleList
	default:
		return vk.PrimitiveTopologyTriangleList
	}
}
