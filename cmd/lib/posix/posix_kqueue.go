package posix

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Kqueue(handlePtr *uintptr, namePtr Cstring) uintptr {
	return libScePosix_kqueue(handlePtr, namePtr)
}

func libScePosix_kqueue(handlePtr *uintptr, namePtr Cstring) uintptr {
	if handlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("kqueue"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	equeue := CreateEqueue("unnamed")
	var name string
	if namePtr != nil {
		name = GoString(namePtr)
	}
	if name == "" {
		name = fmt.Sprintf("0x%X", equeue.Handle)
	}
	equeue.Name = name
	*handlePtr = equeue.Handle

	logger.Printf("%-132s %s created equeue %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("kqueue"),
		color.Yellow.Sprintf("0x%X", equeue.Handle),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}
