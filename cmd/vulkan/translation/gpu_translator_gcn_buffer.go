package translation

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetBufferFromAddress(address uintptr) (vk.Buffer, uintptr, error) {
	if address >= GlobalGpuAllocator.Base && address < GlobalGpuAllocator.Base+uintptr(GlobalGpuAllocator.Size) {
		return GlobalGpuAllocator.Buffer, address - GlobalGpuAllocator.Base, nil
	}
	if address >= GlobalAllocator.Base && address < GlobalAllocator.Base+uintptr(GlobalAllocator.Size) {
		return GlobalAllocator.Buffer, address - GlobalAllocator.Base, nil
	}
	return vk.NullBuffer, 0, fmt.Errorf("address 0x%X is out of known memory bounds", address)
}
