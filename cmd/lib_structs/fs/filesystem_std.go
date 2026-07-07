package fs

import (
	"errors"
	"fmt"
	ioFs "io/fs"
	"strings"

	"github.com/gookit/color"
)

var OutputPrintf func(message string)
var OutputPrintln func()

type OutputDevice struct {
	Name  string
	Color color.Color
}

func (out *OutputDevice) Read(b []byte) (int, error) {
	return 0, errors.New("output device read not implemented")
}

func (out *OutputDevice) Write(b []byte) (int, error) {
	message := string(b)
	OutputPrintf(fmt.Sprintf("%s %s",
		color.Magenta.Sprintf("[write on %s]", out.Name),
		out.Color.Sprint(message),
	))
	if !strings.HasSuffix(message, "\n") {
		OutputPrintln()
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

func (shFs *SharkieFilesystem) InitializeStdDevices() error {
	if _, err := shFs.Create(GetUsablePath("stdin")); err != nil {
		return err
	}

	stdoutDevice := &OutputDevice{Name: "stdout", Color: color.White}
	shFs.Devices[GetUsablePath("stdout")] = func() PosixFile {
		return stdoutDevice
	}
	if _, err := shFs.Create(GetUsablePath("stdout")); err != nil {
		panic(err)
	}

	stderrDevice := &OutputDevice{Name: "stderr", Color: color.Red}
	shFs.Devices[GetUsablePath("stderr")] = func() PosixFile {
		return stderrDevice
	}
	if _, err := shFs.Create(GetUsablePath("stderr")); err != nil {
		panic(err)
	}

	consoleDevice := &OutputDevice{Name: "/dev/console", Color: color.Cyan}
	shFs.Devices[GetUsablePath("/dev/console")] = func() PosixFile {
		return consoleDevice
	}
	if _, err := shFs.Create(GetUsablePath("/dev/console")); err != nil {
		panic(err)
	}

	ttyDevice := &OutputDevice{Name: "/dev/deci_tty6", Color: color.Cyan}
	shFs.Devices[GetUsablePath("/dev/deci_tty6")] = func() PosixFile {
		return ttyDevice
	}
	if _, err := shFs.Create(GetUsablePath("/dev/deci_tty6")); err != nil {
		panic(err)
	}

	return nil
}
