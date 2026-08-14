package ssl

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

var NextSslId = uint32(0x1001)

// 0x0000000000002F40
// __int64 __fastcall sceSslInit(__int64)
func libSceSsl_sceSslInit(poolSize uint64) uintptr {
	id := NextSslId
	NextSslId++

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceSslInit"),
		color.Yellow.Sprintf("0x%X", id),
	)
	return uintptr(id)
}
