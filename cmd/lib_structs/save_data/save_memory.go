package save_data

import (
	"os"
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/config"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
)

type SaveMemory struct {
	UserId     UserId
	GameSerial string
	DirName    string
	MemorySize uint64
	Data       []byte

	ParamSfo   *Psf
	MountPoint string
}

func NewSaveMemory(userId UserId, gameSerial, dirName string, memorySize uint64) *SaveMemory {
	return &SaveMemory{
		UserId:     userId,
		GameSerial: gameSerial,
		DirName:    dirName,
		MemorySize: memorySize,
		Data:       []byte{},
	}
}

func (memory *SaveMemory) ExistsOnHost() bool {
	if _, err := os.Stat(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys", "param.sfo")); err == nil {
		return true
	}

	return false
}

func (memory *SaveMemory) Load(initIconData []byte, initParam *SaveDataParam) (error, bool) {
	var err error
	created := false
	if !memory.ExistsOnHost() {
		memory.ParamSfo, err = memory.CreateOnHost(initIconData, initParam)
		if err != nil {
			return err, false
		}
		created = true
	} else {
		memory.ParamSfo, err = NewPsfFromPath(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys", "param.sfo"))
		if err != nil {
			return err, false
		}
	}

	// Load data.
	memoryPath := filepath.Join(config.GetGameSaveDir(memory.DirName), "memory.dat")
	if _, err = os.Stat(memoryPath); err == nil {
		data, err := os.ReadFile(memoryPath)
		if err != nil {
			return err, false
		}
		memory.Data = data
	}

	return nil, created
}

func (memory *SaveMemory) Save() error {
	_, err := os.Stat(filepath.Join(config.GetGameSaveDir(memory.DirName), "memory.dat"))
	_ = err

	return nil
}

func (memory *SaveMemory) CreateOnHost(initIconData []byte, initParam *SaveDataParam) (*Psf, error) {
	psf, err := NewDefaultParamSfo(memory.DirName, memory.GameSerial)
	if initParam != nil {
		initParam.SaveToParamSfo(psf)
	}
	if err != nil {
		return nil, err
	}

	data, err := psf.Encode()
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys"), 0755); err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys", "param.sfo"), data, 0755); err != nil {
		return nil, err
	}
	if initIconData != nil {
		if err = os.WriteFile(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys", "save_data.png"), initIconData, 0755); err != nil {
			return nil, err
		}
	}

	return psf, nil
}

func (memory *SaveMemory) SaveParamSfo() error {
	if memory.ParamSfo == nil {
		return nil
	}
	data, err := memory.ParamSfo.Encode()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(config.GetGameSaveDir(memory.DirName), "sce_sys", "param.sfo"), data, 0755)
}
