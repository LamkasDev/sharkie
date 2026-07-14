package save_data

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/config"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
	"go101.org/nstd"
)

const (
	SaveParamAccountID         = "ACCOUNT_ID"
	SaveParamMainTitle         = "MAINTITLE"
	SaveParamSubtitle          = "SUBTITLE"
	SaveParamDetail            = "DETAIL"
	SaveParamSaveDataDirectory = "SAVEDATA_DIRECTORY"
	SaveParamSaveDataListParam = "SAVEDATA_LIST_PARAM"
	SaveParamTitleID           = "TITLE_ID"
	SaveParamSaveDataBlocks    = "SAVEDATA_BLOCKS"

	OrbisSaveDataBlockSize = uint64(32768)
	OrbisSaveDataBlocksMin = uint64(96)
	OrbisSaveDataBlocksMax = math.MaxUint64
)

type SaveInstance struct {
	UserId     int32
	GameSerial string
	DirName    string
	MaxBlocks  uint64

	ParamSfo   *Psf
	MountPoint string
}

func NewSaveInstance(userId int32, gameSerial, dirName string, maxBlocks uint64) *SaveInstance {
	maxBlocks = nstd.Clamp(maxBlocks, OrbisSaveDataBlocksMin, OrbisSaveDataBlocksMax)
	return &SaveInstance{
		UserId:     userId,
		GameSerial: gameSerial,
		DirName:    dirName,
		MaxBlocks:  maxBlocks,
	}
}

func (instance *SaveInstance) ExistsOnHost() bool {
	if _, err := os.Stat(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys", "param.sfo")); err == nil {
		return true
	}

	return false
}

func (instance *SaveInstance) Mounted() bool {
	_, err := GlobalFilesystem.Stat(fmt.Sprintf("%s/sce_sys/param.sfo", instance.MountPoint))
	if err != nil {
		return false
	}

	return true
}

// TODO: readOnly, ignoreCorrupt, dontRestoreBackup
func (instance *SaveInstance) Mount(mountSlot int, copyIcon bool) (error, bool) {
	var err error
	created := false
	if !instance.ExistsOnHost() {
		instance.ParamSfo, err = instance.CreateOnHost(copyIcon)
		if err != nil {
			return err, false
		}
		created = true
	} else {
		instance.ParamSfo, err = NewPsfFromPath(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys", "param.sfo"))
		if err != nil {
			return err, false
		}
	}
	instance.MaxBlocks = GetMaxBlocksFromSfo(instance.ParamSfo)

	// Mount save.
	instance.MountPoint = fmt.Sprintf("/savedata%d", mountSlot)
	if err = GlobalFilesystem.Mount(instance.MountPoint, config.GetGameSaveDir(instance.DirName), false); err != nil {
		return err, false
	}

	return nil, created
}

func (instance *SaveInstance) CreateOnHost(copyIcon bool) (*Psf, error) {
	psf, err := NewDefaultParamSfo(instance.DirName, instance.GameSerial)
	if err != nil {
		return nil, err
	}
	if err = psf.AddBinaryUint64(SaveParamSaveDataBlocks, instance.MaxBlocks, true); err != nil {
		return nil, err
	}

	data, err := psf.Encode()
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys"), 0755); err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys", "param.sfo"), data, 0755); err != nil {
		return nil, err
	}
	if copyIcon {
		if iconData, err := GlobalFilesystem.ReadFull("/app0/sce_sys/save_data.png"); err == nil {
			if err = os.WriteFile(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys", "save_data.png"), iconData, 0755); err != nil {
				return nil, err
			}
		}
	}

	return psf, nil
}

func NewDefaultParamSfo(dirName, gameSerial string) (*Psf, error) {
	psf := NewPsf()
	if err := psf.AddBinary(SaveParamAccountID, make([]byte, 8), false); err != nil {
		return nil, err
	}
	if err := psf.AddString(SaveParamMainTitle, "Saved Data", false); err != nil {
		return nil, err
	}
	if err := psf.AddString(SaveParamSubtitle, "", false); err != nil {
		return nil, err
	}
	if err := psf.AddString(SaveParamDetail, "", false); err != nil {
		return nil, err
	}
	if err := psf.AddString(SaveParamSaveDataDirectory, dirName, false); err != nil {
		return nil, err
	}
	if err := psf.AddInteger(SaveParamSaveDataListParam, 0, false); err != nil {
		return nil, err
	}
	if err := psf.AddString(SaveParamTitleID, gameSerial, false); err != nil {
		return nil, err
	}
	if err := psf.AddBinaryUint64(SaveParamSaveDataBlocks, 0, false); err != nil {
		return nil, err
	}

	return psf, nil
}

func GetMaxBlocksFromSfo(psf *Psf) uint64 {
	value, ok := psf.GetBinary(SaveParamSaveDataBlocks)
	if !ok || len(value) < 8 {
		return OrbisSaveDataBlocksMax
	}

	return binary.LittleEndian.Uint64(value)
}

func (instance *SaveInstance) SaveParamSfo() error {
	if instance.ParamSfo == nil {
		return nil
	}
	data, err := instance.ParamSfo.Encode()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(config.GetGameSaveDir(instance.DirName), "sce_sys", "param.sfo"), data, 0755)
}
