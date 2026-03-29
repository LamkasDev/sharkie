package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/dce"
	"github.com/LamkasDev/sharkie/cmd/structs/gc"
	"github.com/LamkasDev/sharkie/cmd/structs/rng"
	"github.com/gookit/color"
)

var GlobalFilesystem *SharkieFilesystem

var FileDescriptorColors = map[string]color.Color{
	"stdin":          color.White,
	"stdout":         color.White,
	"stderr":         color.Red,
	"/dev/console":   color.Cyan,
	"/dev/deci_tty6": color.Cyan,
}

type DeviceFileCreateFunc func() PosixFile

type SharkieFilesystem struct {
	Descriptors    map[FileDescriptor]*SharkieFile
	NextDescriptor FileDescriptor
	Fs             *FS
	Devices        map[string]DeviceFileCreateFunc
	Lock           sync.Mutex
}

func (shFs *SharkieFilesystem) Open(path string, oflag uintptr, mode uintptr) (FileDescriptor, error) {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	if createFunc, isDevice := shFs.Devices[path]; isDevice {
		fd := shFs.NextDescriptor
		shFs.Descriptors[fd] = &SharkieFile{
			Path:       path,
			Descriptor: fd,
			File:       createFunc(),
		}
		shFs.NextDescriptor++

		return fd, nil
	}

	flag := os.O_RDWR
	if (oflag & SCE_O_CREAT) != 0 {
		flag |= os.O_CREATE
		if err := shFs.Fs.MkdirAll(filepath.ToSlash(filepath.Dir(path)), 0777); err != nil {
			return -1, err
		}
	}

	file, err := shFs.Fs.OpenFile(path, flag, 0777)
	if err != nil {
		return -1, err
	}

	fd := shFs.NextDescriptor
	shFs.Descriptors[fd] = &SharkieFile{
		Path:       path,
		Descriptor: fd,
		File:       file,
	}
	shFs.NextDescriptor++

	return fd, nil
}

func (shFs *SharkieFilesystem) AllocateFd(path string, file PosixFile) FileDescriptor {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	fd := shFs.NextDescriptor
	shFs.Descriptors[fd] = &SharkieFile{
		Path:       path,
		Descriptor: fd,
		File:       file,
	}
	shFs.NextDescriptor++

	return fd
}

func (shFs *SharkieFilesystem) Create(path string) (FileDescriptor, error) {
	return shFs.Open(path, SCE_O_CREAT, 0)
}

func (shFs *SharkieFilesystem) Write(path string, data []byte) (int, error) {
	if err := shFs.Fs.MkdirAll(filepath.ToSlash(filepath.Dir(path)), 0777); err != nil {
		return 0, err
	}
	file, err := shFs.Fs.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Write(data)
}

func (shFs *SharkieFilesystem) WriteFd(fd FileDescriptor, data []byte) (int, error) {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	shFile, ok := shFs.Descriptors[fd]
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.File.Write(data)
}

func (shFs *SharkieFilesystem) ReadFull(path string) ([]byte, error) {
	return shFs.Fs.ReadFile(path)
}

func (shFs *SharkieFilesystem) ReadFullFd(fd FileDescriptor) ([]byte, error) {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	shFile, ok := shFs.Descriptors[fd]
	if !ok {
		return nil, errors.New("invalid file descriptor")
	}

	return io.ReadAll(shFile.File)
}

func (shFs *SharkieFilesystem) Read(path string, data []byte) (int, error) {
	file, err := shFs.Fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Read(data)
}

func (shFs *SharkieFilesystem) ReadFd(fd FileDescriptor, data []byte) (int, error) {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	shFile, ok := shFs.Descriptors[fd]
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.File.Read(data)
}

func (shFs *SharkieFilesystem) Close(fd FileDescriptor) error {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	shFile, ok := shFs.Descriptors[fd]
	if !ok {
		return errors.New("invalid file descriptor")
	}
	if err := shFile.File.Close(); err != nil {
		return err
	}
	delete(shFs.Descriptors, fd)

	return nil
}

func (shFs *SharkieFilesystem) Delete(path string) error {
	return shFs.Fs.Remove(path)
}

func SetupFilesystem() {
	GlobalFilesystem = NewFilesystem()
}

func NewFilesystem() *SharkieFilesystem {
	shFs := &SharkieFilesystem{
		Descriptors:    map[FileDescriptor]*SharkieFile{},
		NextDescriptor: 0x0,
		Fs:             NewFS(),
		Devices:        map[string]DeviceFileCreateFunc{},
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
	// Device files.
	if _, err := shFs.Create(GetUsablePath("stdin")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("stdout")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("stderr")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/console")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/deci_tty6")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/dipsw")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_cmd")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_snsr")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_3da")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_dist")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/sbl_srv")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hid")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/ajm")); err != nil {
		return err
	}

	// Actual devices.
	shFs.Devices[GetUsablePath("/dev/rng")] = func() PosixFile {
		return rng.GlobalRngDevice
	}
	shFs.Devices[GetUsablePath("/dev/gc")] = func() PosixFile {
		return gc.GlobalGraphicsController
	}
	shFs.Devices[GetUsablePath("/dev/dce")] = func() PosixFile {
		return dce.GlobalDisplayCoreEngine
	}

	// Deamon files.
	if _, err := shFs.Write(GetUsablePath(structs.AudioInBufferName), make([]byte, structs.AudioInBufferDefault)); err != nil {
		panic(err)
	}
	if _, err := shFs.Write(GetUsablePath(structs.AudioVideoSettingsName), make([]byte, structs.AudioVideoSettingsDefault)); err != nil {
		panic(err)
	}
	if _, err := shFs.Write(GetUsablePath("SceNpTpip"), make([]byte, 4096)); err != nil {
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
		fsPath = GetUsablePath(filepath.ToSlash(fsPath))
		written, err := shFs.Write(fsPath, data)
		if err != nil {
			return err
		}
		logger.Printf(
			"Loaded file %s as %s (size=%s).\n",
			color.Blue.Sprint(path),
			color.Blue.Sprint(fsPath),
			color.Yellow.Sprintf("0x%X", written),
		)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
