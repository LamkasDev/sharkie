package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const (
	RTP_LOOKUP        = 0
	RTP_PRIO_NORMAL   = 1
	RTP_PRIO_REALTIME = 2
)

// 0x0000000000001710
// __int64 __fastcall rtprio_thread()
func libKernel_rtprio_thread(function, lwpid, rtpPtr uintptr) uintptr {
	if rtpPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid structs pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("rtprio_thread"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	switch function {
	case RTP_LOOKUP:
		rtpSlice := unsafe.Slice((*byte)(unsafe.Pointer(rtpPtr)), 4)
		binary.LittleEndian.PutUint16(rtpSlice, RTP_PRIO_NORMAL)
		binary.LittleEndian.PutUint16(rtpSlice[2:], 0)
		logger.Printf("%-132s %s requested rtp struct (type=%s, priority=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("rtprio_thread"),
			color.Yellow.Sprintf("0x%X", RTP_PRIO_NORMAL),
			color.Yellow.Sprintf("0x%X", 0),
		)
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown method %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("rtprio_thread"),
		color.Yellow.Sprintf("0x%X", function),
	)
	SetErrno(EINVAL)
	return ERR_PTR
}
