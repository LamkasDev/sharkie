package np_common

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterNpCommonStubs() {
	// Setup functions.
	elf.RegisterStub("libSceNpCommon", "sceNpCmpNpId", libSceNpCommon_sceNpCmpNpId)
}
