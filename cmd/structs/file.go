package structs

import (
	"io"
	"io/fs"
	"strings"
)

const (
	SCE_O_CREAT = 0x200
)

type SharkieFile struct {
	Path       string
	Descriptor int32
	File       fs.File
}

func GetUsablePath(path string) string {
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return "unnamed"
	}

	return path
}

func (shFile *SharkieFile) Read(data []byte) (int, error) {
	return shFile.File.Read(data)
}

func (shFile *SharkieFile) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := shFile.File.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	panic("failed shared memory file seek!")
}

func (shFile *SharkieFile) Close() error {
	return shFile.File.Close()
}
