package msg_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
)

func RegisterMsgDialogStubs() {
	elf.RegisterStub("libSceMsgDialog", "sceMsgDialogInitialize", SceMsgDialogInitialize)
	elf.RegisterStub("libSceMsgDialog", "sceMsgDialogOpen", SceMsgDialogOpen)
	elf.RegisterStub("libSceMsgDialog", "sceMsgDialogUpdateStatus", SceMsgDialogUpdateStatus)
	elf.RegisterStub("libSceMsgDialog", "sceMsgDialogGetResult", SceMsgDialogGetResult)
	elf.RegisterStub("libSceMsgDialog", "sceMsgDialogTerminate", SceMsgDialogTerminate)
}

func SceMsgDialogInitialize() uintptr {
	logger.Printf("sceMsgDialogInitialize called\n")
	return 0
}

func SceMsgDialogOpen(paramPtr uintptr) uintptr {
	logger.Printf("sceMsgDialogOpen called (paramPtr=%x)\n", paramPtr)
	return 0
}

func SceMsgDialogUpdateStatus() uintptr {
	logger.Printf("sceMsgDialogUpdateStatus called\n")
	return 0
}

func SceMsgDialogGetResult(resultPtr uintptr) uintptr {
	logger.Printf("sceMsgDialogGetResult called (resultPtr=%x)\n", resultPtr)
	return 0
}

func SceMsgDialogTerminate() uintptr {
	logger.Printf("sceMsgDialogTerminate called\n")
	return 0
}
