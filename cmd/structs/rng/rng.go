package rng

import (
	"errors"
	ioFs "io/fs"
	"math/rand"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

var GlobalRngDevice *RngDevice

const (
	SCE_RNG_IOCTL_GET_ENTROPY = 0x40445302
)

type RngDevice struct {
	Rand *rand.Rand
}

func NewRngDevice() *RngDevice {
	return &RngDevice{
		Rand: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (s *RngDevice) Read(b []byte) (int, error) {
	return 0, errors.New("rng read not implemented")
}

func (s *RngDevice) Write(b []byte) (int, error) {
	return 0, errors.New("rng write not implemented")
}

func (s *RngDevice) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("rng seek not implemented")
}

func (s *RngDevice) Close() error {
	return nil
}

func (s *RngDevice) Stat() (ioFs.FileInfo, error) {
	return nil, errors.New("rng stat not implemented")
}

func (s *RngDevice) Truncate(size int64) error {
	return errors.New("rng truncate not implemented")
}

func (s *RngDevice) Ioctl(request uint64, argPtr uintptr) error {
	switch request {
	case SCE_RNG_IOCTL_GET_ENTROPY:
		size := (request >> 16) & 0x1FFF
		argSlice := unsafe.Slice((*byte)(unsafe.Pointer(argPtr)), size)
		if _, err := GlobalRngDevice.Rand.Read(argSlice); err != nil {
			return err
		}

		logger.Printf("%-132s %s wrote %s random bytes to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", size),
			color.Yellow.Sprintf("0x%X", argPtr),
		)
		return nil
	}

	return errors.New("unknown rng ioctl")
}

func SetupRngDevice() {
	GlobalRngDevice = NewRngDevice()
	if _, err := fs.GlobalFilesystem.Create(fs.GetUsablePath("/dev/rng")); err != nil {
		panic(err)
	}
	fs.GlobalFilesystem.Devices[fs.GetUsablePath("/dev/rng")] = func() fs.PosixFile {
		return GlobalRngDevice
	}
}
