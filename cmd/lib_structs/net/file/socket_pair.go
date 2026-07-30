package file

import (
	"bytes"
	"errors"
	"io/fs"
	"sync"
)

type SocketPair struct {
	Name     string
	Domain   int32
	Type     int32
	Protocol int32

	Peer   *SocketPair
	Buffer bytes.Buffer
	Mutex  sync.Mutex
	Closed bool
}

func (s *SocketPair) Read(b []byte) (int, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.Buffer.Len() == 0 {
		if s.Closed {
			return 0, errors.New("socket closed")
		}
		return 0, nil
	}

	return s.Buffer.Read(b)
}

func (s *SocketPair) Write(b []byte) (int, error) {
	if s.Peer == nil {
		return 0, errors.New("broken pipe")
	}
	s.Peer.Mutex.Lock()
	defer s.Peer.Mutex.Unlock()
	if s.Peer.Closed {
		return 0, errors.New("broken pipe")
	}

	return s.Peer.Buffer.Write(b)
}

func (s *SocketPair) Close() error {
	s.Mutex.Lock()
	s.Closed = true
	s.Buffer.Reset()
	s.Mutex.Unlock()

	return nil
}

func (s *SocketPair) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("socket seek not implemented")
}

func (s *SocketPair) Stat() (fs.FileInfo, error) {
	return nil, errors.New("socket stat not implemented")
}

func (s *SocketPair) Truncate(size int64) error {
	return errors.New("socket truncate not implemented")
}

func (s *SocketPair) Ioctl(request uint64, argPtr uintptr) error {
	return errors.New("unknown socket ioctl")
}
