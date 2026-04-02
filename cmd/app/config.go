package app

import (
	"path"
	"time"
)

// Config holds startup configuration for the application.
type Config struct {
	MonitorNum   int
	WindowWidth  uint32
	WindowHeight uint32
	RefreshRate  time.Duration

	IconPath string
	FontPath string

	DebugMode bool
}

func DefaultConfig() Config {
	return Config{
		MonitorNum:   0,
		WindowWidth:  800,
		WindowHeight: 600,
		RefreshRate:  time.Second / 60,
		IconPath:     path.Join("winres", "icon.png"),
		FontPath:     path.Join("data", "JetBrainsMono-Regular.ttf"),
		DebugMode:    true,
	}
}
