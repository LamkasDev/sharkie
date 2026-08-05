package video_out

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000C530
// __int64 __fastcall sceVideoOutGetEventId(__int64)
func libSceVideoOut_sceVideoOutGetEventId(event *KernelEvent) uintptr {
	if event == nil {
		logger.Printf("%-132s %s failed due to invalid event pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventId"),
		)
		return 0x80290002
	}
	if event.Filter != KernelEventFilterVideoOut {
		logger.Printf("%-132s %s failed due to invalid event filter.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventId"),
		)
		return 0x8029000D
	}

	switch event.Id {
	case VideoOutInternalEventIdFlip:
		return VideoOutEventIdFlip
	case VideoOutInternalEventIdVblank, VideoOutInternalEventIdSysVblank:
		return VideoOutEventIdVblank
	case VideoOutInternalEventIdPreVblankStart:
		return VideoOutEventIdPreVblankStart
	case VideoOutInternalEventIdSetMode:
		return VideoOutEventIdSetMode
	case VideoOutInternalEventIdPosition:
		return VideoOutEventIdPosition
	}

	logger.Printf("%-132s %s failed due to invalid event id.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutGetEventId"),
	)
	return 0x8029000D
}

// 0x000000000000C5D0
// __int64 __fastcall sceVideoOutGetEventData(__int64, unsigned __int64 *)
func libSceVideoOut_sceVideoOutGetEventData(event *KernelEvent, data *uint64) uintptr {
	if event == nil || data == nil {
		logger.Printf("%-132s %s failed due to invalid event or data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventData"),
		)
		return 0x80290002
	}
	if event.Filter != KernelEventFilterVideoOut {
		logger.Printf("%-132s %s failed due to invalid event filter.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventData"),
		)
		return 0x8029000D
	}

	eventData := event.FilterData >> 10
	if event.Id != VideoOutInternalEventIdFlip || event.FilterData >= 0 {
		*data = eventData
	} else {
		*data = eventData | 0xFFFF0000_00000000
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s returned event data.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventData"),
		)
	}
	return 0
}

// 0x000000000000C5A0
// __int64 __fastcall sceVideoOutGetEventCount(__int64)
func libSceVideoOut_sceVideoOutGetEventCount(event *KernelEvent) uintptr {
	if event == nil {
		logger.Printf("%-132s %s failed due to invalid event pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventCount"),
		)
		return 0x80290002
	}
	if event.Filter != KernelEventFilterVideoOut {
		logger.Printf("%-132s %s failed due to invalid event filter.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetEventCount"),
		)
		return 0x8029000D
	}

	return uintptr((event.FilterData >> 16) & 0xFFFFFFFFFFFF)
}
