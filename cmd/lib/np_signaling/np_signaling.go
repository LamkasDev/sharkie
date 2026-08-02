package np_signaling

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterNpSignalingStubs() {
	elf.RegisterStub("libSceNpSignaling", "sceNpSignalingInitialize", libSceNpSignaling_sceNpSignalingInitialize)
}
