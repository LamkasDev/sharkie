package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
)

type GameInfo struct {
	Path     string
	Title    string
	TitleId  string
	IconPath *string
}

// GetLibraryGames scans a list of directories and returns a list of games.
func GetLibraryGames(libraryDirs []string) ([]GameInfo, error) {
	var games []GameInfo
	for _, libraryDir := range libraryDirs {
		entries, err := os.ReadDir(libraryDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			gamePath := filepath.Join(libraryDir, entry.Name())
			paramSfoPath := filepath.Join(gamePath, "Sc0", "param.sfo")
			if _, err = os.Stat(paramSfoPath); err != nil {
				continue
			}
			psfData, err := psf.NewPsfFromPath(paramSfoPath)
			if err != nil {
				fmt.Printf("failed to read %s: %v\n", paramSfoPath, err)
				continue
			}
			gameInfo := GameInfo{
				Path: gamePath,
			}
			gameInfo.Title, _ = psfData.MapStrings["TITLE"]
			gameInfo.TitleId, _ = psfData.MapStrings["TITLE_ID"]
			iconPath := filepath.Join(gamePath, "Sc0", "icon0.png")
			if _, err := os.Stat(iconPath); err == nil {
				gameInfo.IconPath = &iconPath
			}
			games = append(games, gameInfo)
		}
	}

	return games, nil
}
