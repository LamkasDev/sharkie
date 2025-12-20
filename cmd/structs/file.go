package structs

import "github.com/gookit/color"

const (
	// Standard POSIX
	FD_STDIN  = uintptr(0)
	FD_STDOUT = uintptr(1)
	FD_STDERR = uintptr(2)

	// PS4 Specific
	FD_CONSOLE = uintptr(10)
	FD_TTY     = uintptr(11)
	FD_GC      = uintptr(12)
	FD_DIPSW   = uintptr(13)
)

var FileDescriptors = map[string]uintptr{
	"stdin":          FD_STDIN,
	"stdout":         FD_STDOUT,
	"stderr":         FD_STDERR,
	"/dev/console":   FD_CONSOLE,
	"/dev/deci_tty6": FD_TTY,
	"/dev/gc":        FD_GC,
	"/dev/dipsw":     FD_DIPSW,
}

var FileDescriptorNames = map[uintptr]string{
	FD_STDIN:   "stdin",
	FD_STDOUT:  "stdout",
	FD_STDERR:  "stderr",
	FD_CONSOLE: "/dev/console",
	FD_TTY:     "/dev/deci_tty6",
	FD_GC:      "/dev/gc",
	FD_DIPSW:   "/dev/dipsw",
}

var FileDescriptorColors = map[uintptr]color.Color{
	FD_STDIN:   color.White,
	FD_STDOUT:  color.White,
	FD_STDERR:  color.Red,
	FD_CONSOLE: color.Cyan,
	FD_TTY:     color.Cyan,
	FD_GC:      color.Cyan,
	FD_DIPSW:   color.Cyan,
}
