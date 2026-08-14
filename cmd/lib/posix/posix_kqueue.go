package posix

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Kqueue() uintptr {
	return libScePosix_kqueue()
}

func libScePosix_kqueue() uintptr {
	equeue := CreateEqueue("unnamed")
	equeue.Name = fmt.Sprintf("0x%X", equeue.Handle)

	logger.Printf("%-132s %s created equeue %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("kqueue"),
		color.Yellow.Sprintf("0x%X", equeue.Handle),
		color.Blue.Sprint(equeue.Name),
	)
	return equeue.Handle
}
