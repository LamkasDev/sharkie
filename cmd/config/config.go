package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	gap "github.com/muesli/go-app-paths"
)

type Config struct {
	LibraryDirectories []string `json:"library_directories"`
	InputMode          string   `json:"input_mode"` // "keyboard" or "controller"
	SyncGuestFlips     bool     `json:"sync_guest_flips"`
}

func NewDefaultConfig() (*Config, error) {
	dataDir, err := AppScope.DataPath("library")
	if err != nil {
		return nil, err
	}

	return &Config{
		LibraryDirectories: []string{dataDir},
		InputMode:          "keyboard",
	}, nil
}

var (
	AppScope      *gap.Scope
	GlobalConfig  *Config
	GameName      string
	GameDirectory string
)

func init() {
	AppScope = gap.NewScope(gap.User, "sharkie")
	err := os.MkdirAll(GetLibDir(), 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(GetToolsDir(), 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(GetSavesDir(), 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(GetAddonsDir(), 0755)
	if err != nil {
		panic(err)
	}
}

func LoadConfig() error {
	configPath, err := AppScope.ConfigPath("config.json")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		GlobalConfig, err = NewDefaultConfig()
		if err != nil {
			return err
		}
		if err = SaveConfig(); err != nil {
			return err
		}
		return nil
	}

	GlobalConfig = &Config{}
	if err = json.Unmarshal(data, GlobalConfig); err != nil {
		return err
	}

	return nil
}

func SaveConfig() error {
	configPath, err := AppScope.ConfigPath("config.json")
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func GetGameCacheDir() string {
	cachePath, _ := AppScope.CacheDir()
	return filepath.Join(cachePath, GameName)
}

func GetGameSavesDir() string {
	return filepath.Join(GetSavesDir(), GameName)
}

func GetGameSaveDir(dirName string) string {
	return filepath.Join(GetGameSavesDir(), dirName)
}

func GetGameAddonsDir(titleId string) string {
	return filepath.Join(GetAddonsDir(), titleId)
}

func GetLibDir() string {
	libPath, _ := AppScope.DataPath("lib")
	return libPath
}

func GetToolsDir() string {
	toolsPath, _ := AppScope.DataPath("tools")
	return toolsPath
}

func GetSavesDir() string {
	savesPath, _ := AppScope.DataPath("saves")
	return savesPath
}

func GetAddonsDir() string {
	addonsPath, _ := AppScope.DataPath("addons")
	return addonsPath
}

func ResolveGame(arg string) error {
	// Try as an absolute or relative directory first.
	info, err := os.Stat(arg)
	if err == nil && info.IsDir() {
		if _, err = os.Stat(filepath.Join(arg, "Image0", "eboot.bin")); err == nil {
			GameDirectory = arg
			GameName = filepath.Base(arg)
			return nil
		}
		if _, err = os.Stat(filepath.Join(arg, "Image0", "eboot.elf")); err == nil {
			GameDirectory = arg
			GameName = filepath.Base(arg)
			return nil
		}
	}

	// Not a valid direct path, try searching in LibraryDirectories.
	if GlobalConfig != nil {
		for _, libDir := range GlobalConfig.LibraryDirectories {
			gamePath := filepath.Join(libDir, arg)
			info, err = os.Stat(gamePath)
			if err == nil && info.IsDir() {
				if _, err = os.Stat(filepath.Join(gamePath, "Image0", "eboot.bin")); err == nil {
					GameDirectory = gamePath
					GameName = filepath.Base(gamePath)
					return nil
				}
				if _, err = os.Stat(filepath.Join(gamePath, "Image0", "eboot.elf")); err == nil {
					GameDirectory = gamePath
					GameName = filepath.Base(gamePath)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("could not resolve game: %s (checked direct path and library directories)", arg)
}
