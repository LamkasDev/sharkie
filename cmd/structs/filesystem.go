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

func (shFs *SharkieFilesystem) Open(path string, oflag uintptr, mode uintptr) (*SharkieFile, error) {
	file, ok := shFs.Files[path]
	if !ok {
		if (oflag & SCE_O_CREAT) != 0 {
			return shFs.Create(path)
		}
		return nil, errors.New("file not found")
	}

	return file, nil
}

func (shFs *SharkieFilesystem) Create(path string) (*SharkieFile, error) {
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

func (shFs *SharkieFilesystem) Write(path string, data []byte) (int, error) {
	_, err := shFs.Open(path, SCE_O_CREAT, 0)
	if err != nil {
		return 0, err
	}
	return len(data), shFs.Fs.WriteFile(GetUsablePath(path), data, 0777)
}

func (shFs *SharkieFilesystem) ReadFull(path string) ([]byte, error) {
	_, err := shFs.Open(path, 0, 0)
	if err != nil {
		return []byte{}, err
	}
	return fs.ReadFile(shFs.Fs, GetUsablePath(path))
}

func (shFs *SharkieFilesystem) Read(path string, data []byte) (int, error) {
	file, err := shFs.Open(path, 0, 0)
	if err != nil {
		return 0, err
	}
	return file.Read(data)
}

func (shFs *SharkieFilesystem) Delete(path string) error {
	file, ok := shFs.Files[path]
	if !ok {
		return errors.New("file not found")
	}
	// I'm not sure if it can be reopened after closing, let's leave it be.
	/* if err := file.Close(); err != nil {
		return err
	} */
	delete(shFs.Files, path)
	delete(shFs.Descriptors, file.Descriptor)

	return nil
}

func NewFilesystem() *SharkieFilesystem {
	fs := &SharkieFilesystem{
		Files:          map[string]*SharkieFile{},
		Descriptors:    map[int32]*SharkieFile{},
		NextDescriptor: 0x0,
		Fs:             memfs.New(),
		Lock:           sync.Mutex{},
	}
	if _, err := fs.Create("stdin"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("stdout"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("stderr"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/console"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/deci_tty6"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/gc"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/dipsw"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/hmd_cmd"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/hmd_snsr"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/hmd_3da"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/hmd_dist"); err != nil {
		panic(err)
	}
	if _, err := fs.Create("/dev/sbl_srv"); err != nil {
		panic(err)
	}
	if _, err := fs.Write(AudioInBufferName, make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}

	return fs
}
