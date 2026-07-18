package app_content

import (
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
)

var GlobalAppContentInstance *AppContentInstance

type AppContentInstance struct {
	IsInitialized bool
	ParamSfo      *psf.Psf
	AddcontInfo   []AppContentAddcontInfo
}

func NewAppContentInstance() *AppContentInstance {
	instance := &AppContentInstance{
		AddcontInfo: []AppContentAddcontInfo{},
	}

	// Load app metadata.
	sfoPath := filepath.Join(config.GameDirectory, "Sc0", "param.sfo")
	p, err := psf.NewPsfFromPath(sfoPath)
	if err != nil {
		panic(err)
	}
	instance.ParamSfo = p
	if titleId := instance.ParamSfo.MapStrings["TITLE_ID"]; titleId == "" {
		panic("missing title id")
	}

	return instance
}

func SetupAppContentInstance() {
	GlobalAppContentInstance = NewAppContentInstance()
}
