package vulkan

import (
	"fmt"
	"os"

	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

var (
	loadingDispatchFile  *os.File
	loadingDmaFile       *os.File
	loadingResourcesFile *os.File
	menuDispatchFile     *os.File
	menuDmaFile          *os.File
	menuDrawFile         *os.File
	menuResourcesFile    *os.File
)

func init() {
	os.MkdirAll("research", 0755)
	loadingDispatchFile, _ = os.OpenFile("research/loading_dispatch_dump.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	loadingDmaFile, _ = os.OpenFile("research/loading_dma_dump.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	loadingResourcesFile, _ = os.OpenFile("research/loading_resources.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	menuDispatchFile, _ = os.OpenFile("research/menu_dispatch_dump.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	menuDmaFile, _ = os.OpenFile("research/menu_dma_dump.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	menuDrawFile, _ = os.OpenFile("research/menu_draw.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	menuResourcesFile, _ = os.OpenFile("research/menu_resources.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func (t *GpuTranslator) ResearchLogDispatch(frame uint64, dispatch *gpu.LiverpoolDispatch) {
	str := fmt.Sprintf("Frame: %d\n", frame)
	str += fmt.Sprintf("Thread groups: X:%d Y:%d Z:%d\n", dispatch.DimX, dispatch.DimY, dispatch.DimZ)
	str += fmt.Sprintf("Compute Shader: 0x%X\n", dispatch.ComputeShaderAddress)
	str += fmt.Sprintf("UserDataHash: 0x%X\n", dispatch.UserDataHash)
	str += "----------------------------------------\n"

	if frame < 300 {
		loadingDispatchFile.WriteString(str)
	} else {
		menuDispatchFile.WriteString(str)
	}
}

func (t *GpuTranslator) ResearchLogDmaCopy(frame uint64, dmaCopy *gpu.LiverpoolDmaCopy) {
	str := fmt.Sprintf("Frame: %d\n", frame)
	str += fmt.Sprintf("Src: 0x%X Dst: 0x%X Count: %d (Size: %d)\n", dmaCopy.SrcAddress, dmaCopy.DstAddress, dmaCopy.Count, dmaCopy.Count*4)
	str += "----------------------------------------\n"

	if frame < 300 {
		loadingDmaFile.WriteString(str)
	} else {
		menuDmaFile.WriteString(str)
	}
}

func (t *GpuTranslator) ResearchLogDraw(frame uint64, draw *gpu.LiverpoolDraw) {
	if frame != 400 {
		return
	}
	str := fmt.Sprintf("Draw Call at Frame 400\n")
	str += fmt.Sprintf("IndexCount: %d\n", draw.IndexCount)
	str += fmt.Sprintf("UserDataHash: 0x%X\n", draw.UserDataHash)
	str += "----------------------------------------\n"
	menuDrawFile.WriteString(str)
}

func (t *GpuTranslator) ResearchDumpResources(frame uint64) {
	if frame != 200 && frame != 400 {
		return
	}

	str := "=== SURFACES ===\n"
	t.surfacesMutex.Lock()
	for key, surface := range t.surfaces {
		str += fmt.Sprintf("GpuAddress: 0x%X Width: %d Height: %d Format: %v FrameUsed: %d\n", key.GpuAddress, surface.Value.width, surface.Value.height, surface.Value.format, surface.FrameUsed)
	}
	t.surfacesMutex.Unlock()

	str += "\n=== TEXTURES ===\n"
	t.imagesMutex.Lock()
	for hash, descriptor := range t.imageDescriptors {
		str += fmt.Sprintf("Hash: 0x%X BaseAddress: 0x%X Width: %d Height: %d Pitch: %d DataFormat: %d NumFormat: %d\n", hash, descriptor.BaseAddress, descriptor.Width, descriptor.Height, descriptor.Pitch, descriptor.DataFormat, descriptor.NumFormat)
	}
	t.imagesMutex.Unlock()

	if frame == 200 {
		loadingResourcesFile.WriteString(str)
	}
	if frame == 400 {
		menuResourcesFile.WriteString(str)
	}
}
