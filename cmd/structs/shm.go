package structs

import (
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/psanford/memfs"
)

var GlobalShmFilesystem = NewShmFilesystem()
var GlobalShmFilesystemLock = sync.Mutex{}

type ShmFilesystem struct {
	Files          map[string]*ShmFile
	Descriptors    map[int32]*ShmFile
	NextDescriptor int32
	Fs             *memfs.FS
}

type ShmFile struct {
	Path       string
	Descriptor int32
	File       fs.File
}

func (shmFs *ShmFilesystem) Open(name string, oflag int32, mode int32) (*ShmFile, error) {
	file, ok := shmFs.Files[name]
	if !ok {
		if (oflag & SCE_O_CREAT) != 0 {
			return shmFs.CreateFile(name)
		}
		return nil, errors.New("file not found")
	}

	return file, nil
}

func (shmFs *ShmFilesystem) CreateFile(name string) (*ShmFile, error) {
	path := GetShmPath(name)
	err := shmFs.Fs.WriteFile(path, []byte{}, 0666)
	if err != nil {
		return nil, err
	}
	file, err := shmFs.Fs.Open(path)
	if err != nil {
		return nil, err
	}
	shmFile := &ShmFile{
		Path:       path,
		Descriptor: shmFs.NextDescriptor,
		File:       file,
	}
	shmFs.Files[name] = shmFile
	shmFs.Descriptors[shmFile.Descriptor] = shmFile
	shmFs.NextDescriptor++

	return shmFile, nil
}

func (shmFs *ShmFilesystem) WriteFile(name string, data []byte) (int, error) {
	file, err := shmFs.Open(name, SCE_O_CREAT, 0)
	if err != nil {
		return 0, err
	}
	return file.Write(data)
}

func (shmFs *ShmFilesystem) ReadFile(name string, data []byte) (int, error) {
	file, err := shmFs.Open(name, 0, 0)
	if err != nil {
		return 0, err
	}
	return file.Read(data)
}

func (shmFile *ShmFile) Read(data []byte) (int, error) {
	return shmFile.File.Read(data)
}

func (shmFile *ShmFile) Write(data []byte) (int, error) {
	return len(data), GlobalShmFilesystem.Fs.WriteFile(shmFile.Path, data, 0777)
}

func (shmFile *ShmFile) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := shmFile.File.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	panic("failed shared memory file seek!")
}

func (shmFile *ShmFile) Close() error {
	return shmFile.File.Close()
}

func GetShmPath(name string) string {
	cleanName := strings.TrimLeft(name, "/")
	cleanName = strings.ReplaceAll(cleanName, "..", "__")
	if cleanName == "" {
		return "default_shm"
	}

	return cleanName
}

func NewShmFilesystem() *ShmFilesystem {
	shmFs := &ShmFilesystem{
		Files:          map[string]*ShmFile{},
		Descriptors:    map[int32]*ShmFile{},
		NextDescriptor: 0x100,
		Fs:             memfs.New(),
	}

	return shmFs
}
