package net

import (
	"errors"
	"io/fs"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

const (
	SCE_NET_IOCTL_INIT = 0x802450C9
)

const SocketBufferSize = 1024 * 1024

type Socket struct {
	Name     string
	Domain   int32
	Type     int32
	Protocol int32
}

func (s *Socket) Read(b []byte) (int, error) {
	return 0, errors.New("socket read not implemented")
}

func (s *Socket) Write(b []byte) (int, error) {
	return 0, errors.New("socket write not implemented")
}

func (s *Socket) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("socket seek not implemented")
}

func (s *Socket) Close() error {
	return nil
}

func (s *Socket) Stat() (fs.FileInfo, error) {
	return nil, errors.New("socket stat not implemented")
}

func (s *Socket) Truncate(size int64) error {
	return errors.New("socket truncate not implemented")
}

func (s *Socket) Ioctl(request uint64, argPtr uintptr) error {
	switch request {
	case SCE_NET_IOCTL_INIT:
		logger.Printf("%-132s %s initialized socket.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return nil
	}

	return errors.New("unknown socket ioctl")
}
