package structs

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/gookit/color"
	"github.com/psanford/memfs"
)

var GlobalFilesystem = NewFilesystem()

var FileDescriptorColors = map[string]color.Color{
	"stdin":          color.White,
	"stdout":         color.White,
	"stderr":         color.Red,
	"/dev/console":   color.Cyan,
	"/dev/deci_tty6": color.Cyan,
	"/dev/gc":        color.Cyan,
	"/dev/dipsw":     color.Cyan,
}

type SharkieFilesystem struct {
	Files          map[string]*SharkieFile
	Descriptors    map[int32]*SharkieFile
	NextDescriptor int32
	Fs             *memfs.FS
	Lock           sync.Mutex
}

func (shFs *SharkieFilesystem) Open(path string, oflag int32, mode int32) (*SharkieFile, error) {
	file, ok := shFs.Files[path]
	if !ok {
		if (oflag & SCE_O_CREAT) != 0 {
			return shFs.CreateFile(path)
		}
		return nil, errors.New("file not found")
	}

	return file, nil
}

func (shFs *SharkieFilesystem) CreateFile(path string) (*SharkieFile, error) {
	err := shFs.Fs.MkdirAll(filepath.Dir(GetUsablePath(path)), 0777)
	if err != nil {
		return nil, err
	}
	err = shFs.Fs.WriteFile(GetUsablePath(path), []byte{}, 0777)
	if err != nil {
		return nil, err
	}
	file, err := shFs.Fs.Open(GetUsablePath(path))
	if err != nil {
		return nil, err
	}
	shmFile := &SharkieFile{
		Path:       path,
		Descriptor: shFs.NextDescriptor,
		File:       file,
	}
	shFs.Files[path] = shmFile
	shFs.Descriptors[shmFile.Descriptor] = shmFile
	shFs.NextDescriptor++

	return shmFile, nil
}

func (shFs *SharkieFilesystem) WriteFile(path string, data []byte) (int, error) {
	_, err := shFs.Open(path, SCE_O_CREAT, 0)
	if err != nil {
		return 0, err
	}
	return len(data), shFs.Fs.WriteFile(GetUsablePath(path), data, 0777)
}

func (shFs *SharkieFilesystem) ReadFullFile(path string) ([]byte, error) {
	_, err := shFs.Open(path, 0, 0)
	if err != nil {
		return []byte{}, err
	}
	return fs.ReadFile(shFs.Fs, GetUsablePath(path))
}

func (shFs *SharkieFilesystem) ReadFile(path string, data []byte) (int, error) {
	file, err := shFs.Open(path, 0, 0)
	if err != nil {
		return 0, err
	}
	return file.Read(data)
}

func NewFilesystem() *SharkieFilesystem {
	fs := &SharkieFilesystem{
		Files:          map[string]*SharkieFile{},
		Descriptors:    map[int32]*SharkieFile{},
		NextDescriptor: 0x0,
		Fs:             memfs.New(),
		Lock:           sync.Mutex{},
	}
	if _, err := fs.CreateFile("stdin"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("stdout"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("stderr"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("/dev/console"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("/dev/deci_tty6"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("/dev/gc"); err != nil {
		panic(err)
	}
	if _, err := fs.CreateFile("/dev/dipsw"); err != nil {
		panic(err)
	}
	if _, err := fs.WriteFile(AudioInBufferName, make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}

	return fs
}
