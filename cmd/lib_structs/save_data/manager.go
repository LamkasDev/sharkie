// Package save_data contains structs to manage save data files.
package save_data

import (
	"fmt"
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

var GlobalSaveDataManager *SaveDataManager

// SaveDataManager keeps track of save mount slots.
type SaveDataManager struct {
	MountSlots [16]*SaveInstance
	Lock       sync.Mutex
}

func NewSaveDataManager() *SaveDataManager {
	return &SaveDataManager{
		MountSlots: [16]*SaveInstance{},
		Lock:       sync.Mutex{},
	}
}

func (sdm *SaveDataManager) GetAvailableMountSlot(dirName string) (int, error) {
	sdm.Lock.Lock()
	defer sdm.Lock.Unlock()
	for slot, save := range sdm.MountSlots {
		if save != nil {
			if save.DirName == dirName {
				return -1, fmt.Errorf("save dir name %s already exists", dirName)
			}
			continue
		}
		return slot, nil
	}

	return -1, fmt.Errorf("no available save mount slot found")
}

func (sdm *SaveDataManager) Unmount(mountPoint string) error {
	sdm.Lock.Lock()
	defer sdm.Lock.Unlock()
	for slot, save := range sdm.MountSlots {
		if save == nil {
			continue
		}
		if save.MountPoint == mountPoint && save.Mounted() {
			if err := save.SaveParamSfo(); err != nil {
				return fmt.Errorf("failed to save param.sfo: %w", err)
			}
			sdm.MountSlots[slot] = nil
			if err := GlobalFilesystem.Unmount(mountPoint); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("no mounted save found")
}

func (sdm *SaveDataManager) GetSaveInstance(mountSlot int) *SaveInstance {
	sdm.Lock.Lock()
	defer sdm.Lock.Unlock()
	return sdm.MountSlots[mountSlot]
}

func (sdm *SaveDataManager) SetSaveInstance(mountSlot int, instance *SaveInstance) {
	sdm.Lock.Lock()
	defer sdm.Lock.Unlock()
	sdm.MountSlots[mountSlot] = instance
}

func SetupSaveDataManager() {
	GlobalSaveDataManager = NewSaveDataManager()
}
