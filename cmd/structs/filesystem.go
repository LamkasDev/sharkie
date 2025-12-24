package structs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
	"github.com/psanford/memfs"
)

var GlobalFilesystem *SharkieFilesystem

var FileDescriptorColors = map[string]color.Color{
	"stdin":          color.White,
	"stdout":         color.White,
	"stderr":         color.Red,
	"/dev/console":   color.Cyan,
	"/dev/deci_tty6": color.Cyan,
}

type SharkieFilesystem struct {
	Files          map[string]*SharkieFile
	Descriptors    map[FileDescriptor]*SharkieFile
	NextDescriptor FileDescriptor
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
	err := shFs.Fs.MkdirAll(filepath.ToSlash(filepath.Dir(GetUsablePath(path))), 0777)
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

func SetupFilesystem() {
	GlobalFilesystem = NewFilesystem()
}

func NewFilesystem() *SharkieFilesystem {
	shFs := &SharkieFilesystem{
		Files:          map[string]*SharkieFile{},
		Descriptors:    map[FileDescriptor]*SharkieFile{},
		NextDescriptor: 0x0,
		Fs:             memfs.New(),
		Lock:           sync.Mutex{},
	}
	if err := shFs.InitializeSystemFiles(); err != nil {
		panic(err)
	}
	if err := shFs.InitializeAppFiles(); err != nil {
		panic(err)
	}

	return shFs
}

func (shFs *SharkieFilesystem) InitializeSystemFiles() error {
	if _, err := shFs.Create("stdin"); err != nil {
		return err
	}
	if _, err := shFs.Create("stdout"); err != nil {
		return err
	}
	if _, err := shFs.Create("stderr"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/console"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/deci_tty6"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/gc"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/dipsw"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/hmd_cmd"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/hmd_snsr"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/hmd_3da"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/hmd_dist"); err != nil {
		return err
	}
	if _, err := shFs.Create("/dev/sbl_srv"); err != nil {
		return err
	}
	if _, err := shFs.Write(AudioInBufferName, make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}

	return nil
}

func (shFs *SharkieFilesystem) InitializeAppFiles() error {
	err := filepath.WalkDir(filepath.Join("fs", "app0"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fsPath, err := filepath.Rel("fs", path)
		if err != nil {
			return err
		}
		fsPath = fmt.Sprintf("/%s", filepath.ToSlash(fsPath))
		_, err = shFs.Write(fsPath, data)
		if err != nil {
			return err
		}
		logger.Printf(
			"Loaded file %s as %s.\n",
			color.Blue.Sprint(path),
			color.Blue.Sprint(fsPath),
		)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
