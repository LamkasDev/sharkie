package app_content

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
)

var GlobalAppContentInstance *AppContentInstance

type AppContentInstance struct {
	IsInitialized bool
	ParamSfo      *psf.Psf
	AddcontInfo   []OrbisAppContentAddcontInfo
}

func NewAppContentInstance() *AppContentInstance {
	return &AppContentInstance{
		AddcontInfo: []OrbisAppContentAddcontInfo{},
	}
}

func SetupAppContentInstance() {
	GlobalAppContentInstance = NewAppContentInstance()
}
