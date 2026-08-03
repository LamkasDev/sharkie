package random

import (
	"math/rand"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterRandomStubs() {
	// Setup functions.
	elf.RegisterStub("libSceRandom", "sceRandomGetRandomNumber", libSceRandom_sceRandomGetRandomNumber)
}

// 0x0000000000000150
// __int64 __fastcall sceRandomGetRandomNumber(__int64, unsigned __int64, __m128 _XMM0)
func libSceRandom_sceRandomGetRandomNumber(bufPtr, size uintptr) uintptr {
	if size > 64 {
		logger.Printf("%-132s %s failed due to exceeding limit.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRandomGetRandomNumber"),
		)
		return 0x817C0016
	}
	buf := unsafe.Slice((*uint8)(unsafe.Pointer(bufPtr)), size)
	for i := range buf {
		buf[i] = uint8(rand.Int())
	}

	logger.Printf("%-132s %s wrote %s random bytes to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceRandomGetRandomNumber"),
		color.Yellow.Sprintf("0x%X", size),
		color.Yellow.Sprintf("0x%X", bufPtr),
	)
	return 0
}
