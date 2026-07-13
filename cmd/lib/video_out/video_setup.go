package video_out

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000AAD0
// __int64 __fastcall sceVideoOutOpen(unsigned int, unsigned int, unsigned int, _DWORD *, __m128 _XMM0)
func libSceVideoOut_sceVideoOutOpen() uintptr {
	handle := &VideoOutHandle{
		Id:                 GlobalDisplayCoreEngine.NextHandle,
		LabelBufferAddress: GlobalGoAllocator.Malloc(uintptr(VideoOutMaxBuffers) * 8),
		NextFlip:           make(chan *VideoOutFlip, VideoOutMaxBuffers),
	}
	GlobalDisplayCoreEngine.Handles[handle.Id] = handle
	GlobalDisplayCoreEngine.NextHandle++

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}
