package fs

import (
	"io"
	"io/fs"
	"strings"
)

const (
	FileBlockSize   = 4096
	MinFileMmapSize = 0x10000
)

type PosixFile interface {
	io.ReadWriteSeeker
	io.Closer
	Stat() (fs.FileInfo, error)
	Truncate(size int64) error
	Ioctl(request uint64, argPtr uintptr) error
}

type FileDescriptor int32

type SharkieFile struct {
	Path       string
	Descriptor FileDescriptor
	File       PosixFile
}

func GetUsablePath(path string) string {
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return "unnamed"
	}

	return path
}

func (shFile *SharkieFile) Read(data []byte) (int, error) {
	return shFile.File.Read(data)
}

func (shFile *SharkieFile) Write(data []byte) (int, error) {
	return shFile.File.Write(data)
}

func (shFile *SharkieFile) Seek(offset int64, whence int) (int64, error) {
	return shFile.File.Seek(offset, whence)
}

func (shFile *SharkieFile) Close() error {
	return shFile.File.Close()
}
