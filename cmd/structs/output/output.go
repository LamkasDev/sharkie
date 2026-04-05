package output

import (
	"errors"
	ioFs "io/fs"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

type OutputDevice struct {
	Name  string
	Color color.Color
}

func (out *OutputDevice) Read(b []byte) (int, error) {
	return 0, errors.New("output device read not implemented")
}

func (out *OutputDevice) Write(b []byte) (int, error) {
	message := string(b)
	logger.Printf("%-132s %s %s",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[write on %s]", out.Name),
		out.Color.Sprint(message),
	)
	if !strings.HasSuffix(message, "\n") {
		logger.Println()
	}

	return len(b), nil
}

func (out *OutputDevice) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("output device seek not implemented")
}

func (out *OutputDevice) Close() error {
	return nil
}

func (out *OutputDevice) Stat() (ioFs.FileInfo, error) {
	return nil, errors.New("output device stat not implemented")
}

func (out *OutputDevice) Truncate(size int64) error {
	return errors.New("output device truncate not implemented")
}

func (out *OutputDevice) Ioctl(request uint64, argPtr uintptr) error {
	return errors.New("unknown output device ioctl")
}

func SetupOutputDevices() {
	stdoutDevice := &OutputDevice{Name: "stdout", Color: color.White}
	if _, err := GlobalFilesystem.Create(GetUsablePath("stdout")); err != nil {
		panic(err)
	}
	GlobalFilesystem.Devices[GetUsablePath("stdout")] = func() PosixFile {
		return stdoutDevice
	}

	stderrDevice := &OutputDevice{Name: "stderr", Color: color.Red}
	if _, err := GlobalFilesystem.Create(GetUsablePath("stderr")); err != nil {
		panic(err)
	}
	GlobalFilesystem.Devices[GetUsablePath("stderr")] = func() PosixFile {
		return stderrDevice
	}

	consoleDevice := &OutputDevice{Name: "/dev/console", Color: color.Cyan}
	if _, err := GlobalFilesystem.Create(GetUsablePath("/dev/console")); err != nil {
		panic(err)
	}
	GlobalFilesystem.Devices[GetUsablePath("/dev/console")] = func() PosixFile {
		return consoleDevice
	}

	ttyDevice := &OutputDevice{Name: "/dev/deci_tty6", Color: color.Cyan}
	if _, err := GlobalFilesystem.Create(GetUsablePath("/dev/deci_tty6")); err != nil {
		panic(err)
	}
	GlobalFilesystem.Devices[GetUsablePath("/dev/deci_tty6")] = func() PosixFile {
		return ttyDevice
	}
}
