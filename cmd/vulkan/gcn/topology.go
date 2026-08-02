package gcn

import vk "github.com/goki/vulkan"

func TranslateTopology(primType uint32) vk.PrimitiveTopology {
	switch primType {
	case 0: // NONE
		panic("unsupported")
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
	case 9, 17: // PATCH + RECTLIST - emulated via tessellation (3 patch verts -> quad)
		return vk.PrimitiveTopologyPatchList
	case 10: // LINELIST_ADJ
		return vk.PrimitiveTopologyLineListWithAdjacency
	case 11: // LINESTRIP_ADJ
		return vk.PrimitiveTopologyLineStripWithAdjacency
	case 12: // TRILIST_ADJ
		return vk.PrimitiveTopologyTriangleListWithAdjacency
	case 13: // TRISTRIP_ADJ
		return vk.PrimitiveTopologyTriangleStripWithAdjacency
	case 16: // TRI_WITH_WFLAGS
		return vk.PrimitiveTopologyTriangleList
	case 18: // LINELOOP
		panic("unsupported")
	case 19: // QUADLIST
		return vk.PrimitiveTopologyTriangleList
	case 20: // QUADSTRIP
		panic("unsupported")
	case 21: // POLYGON
		panic("unsupported")
	case 22: // 2D_COPY_RECT_LIST_V0
		panic("unsupported")
	case 23: // 2D_COPY_RECT_LIST_V1
		panic("unsupported")
	case 24: // 2D_COPY_RECT_LIST_V2
		panic("unsupported")
	case 25: // 2D_COPY_RECT_LIST_V3
		panic("unsupported")
	case 26: // 2D_FILL_RECT_LIST
		panic("unsupported")
	case 27: // 2D_LINE_STRIP
		panic("unsupported")
	case 28: // 2D_TRI_STRIP
		panic("unsupported")
	default:
		panic("unsupported")
	}
}
