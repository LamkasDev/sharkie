package disc_map

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterDiscMapStubs() {
	// Query functions.
	elf.RegisterStub("libSceSceDiscMap", "sceDiscMapGetPackageSize", libSceSceDiscMap_sceDiscMapGetPackageSize)
	elf.RegisterStub("libSceSceDiscMap", "sceDiscMapIsRequestOnHDD", libSceSceDiscMap_sceDiscMapIsRequestOnHDD)
	elf.RegisterStub("libSceSceDiscMap", "fJgP+wqifno#A#B", libSceSceDiscMap_fjg)
}
