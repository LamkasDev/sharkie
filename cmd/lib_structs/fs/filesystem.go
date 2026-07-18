package fs

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/LamkasDev/sharkie/cmd/config"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

var GlobalFilesystem *SharkieFilesystem

type DeviceFileCreateFunc func() PosixFile

// SharkieFilesystem acts as the kernel's file descriptor table.
type SharkieFilesystem struct {
	Descriptors    map[FileDescriptor]*SharkieFile
	NextDescriptor FileDescriptor
	Fs             *FS
	Devices        map[string]DeviceFileCreateFunc
	Lock           sync.Mutex
}

func NewFilesystem() *SharkieFilesystem {
	shFs := &SharkieFilesystem{
		Descriptors:    map[FileDescriptor]*SharkieFile{},
		NextDescriptor: 0x0,
		Fs:             NewFS(),
		Devices:        map[string]DeviceFileCreateFunc{},
		Lock:           sync.Mutex{},
	}
	if err := shFs.InitializeStdDevices(); err != nil {
		panic(err)
	}
	if err := shFs.InitializeSystemFiles(); err != nil {
		panic(err)
	}

	// Mount the game directory as read-only.
	image0Path := filepath.Join(config.GameDirectory, "Image0")
	if err := shFs.Fs.Mount(GetUsablePath("/app0"), image0Path, true); err != nil {
		panic(err)
	}

	return shFs
}

// Create creates a new file and returns the file descriptor.
func (shFs *SharkieFilesystem) Create(path string) (FileDescriptor, error) {
	return shFs.Open(path, O_CREAT|O_RDWR, 0777)
}

// Open creates a new file descriptor for the given path.
func (shFs *SharkieFilesystem) Open(path string, flags FileFlags, mode FileMode) (FileDescriptor, error) {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()

	// Check if it's a registered device file.
	if createFunc, isDevice := shFs.Devices[path]; isDevice {
		fd := shFs.NextDescriptor
		shFs.Descriptors[fd] = &SharkieFile{
			Path:       path,
			Descriptor: fd,
			Device:     createFunc(),
			IsDevice:   true,
			Flags:      int(flags),
		}
		shFs.NextDescriptor++
		return fd, nil
	}

	// Parse basic flags for VFS creation.
	create := (flags & O_CREAT) != 0

	// Resolve or create the node in the VFS tree.
	node, err := shFs.Fs.GetOrCreateNode(path, create, false, os.FileMode(mode))
	if err != nil {
		return -1, err
	}

	// Handle TRUNC flag if it already existed and wasn't a device
	if (flags & O_TRUNC) != 0 {
		if err = node.Truncate(0); err != nil {
			return -1, err
		}
	}

	// Allocate a descriptor.
	fd := shFs.NextDescriptor
	shFile := &SharkieFile{
		Path:       path,
		Descriptor: fd,
		Node:       node,
		Flags:      int(flags),
		Offset:     0,
	}

	// Handle APPEND flag cursor initialization.
	if (flags & O_APPEND) != 0 {
		shFile.Offset = node.GetSize()
	}

	shFs.Descriptors[fd] = shFile
	shFs.NextDescriptor++

	return fd, nil
}

// AllocateFd forcefully maps an existing PosixFile to a new descriptor.
func (shFs *SharkieFilesystem) AllocateFd(path string, file PosixFile) FileDescriptor {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()
	fd := shFs.NextDescriptor
	shFs.Descriptors[fd] = &SharkieFile{
		Path:       path,
		Descriptor: fd,
		Device:     file,
		IsDevice:   true,
	}
	shFs.NextDescriptor++

	return fd
}

func (shFs *SharkieFilesystem) Read(path string, data []byte) (int, error) {
	fd, err := shFs.Open(path, O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer shFs.Close(fd)
	return shFs.ReadFd(fd, data)
}

func (shFs *SharkieFilesystem) ReadFd(fd FileDescriptor, data []byte) (int, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Read(data)
}

func (shFs *SharkieFilesystem) PreadFd(fd FileDescriptor, data []byte, offset int64) (int, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Pread(data, offset)
}

func (shFs *SharkieFilesystem) SeekFd(fd FileDescriptor, offset int64, whence int) (int64, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Seek(offset, whence)
}

func (shFs *SharkieFilesystem) GetOffsetFd(fd FileDescriptor) (int64, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Offset, nil
}

func (shFs *SharkieFilesystem) ReadFull(path string) ([]byte, error) {
	fd, err := shFs.Open(path, O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer shFs.Close(fd)
	return shFs.ReadFullFd(fd)
}

func (shFs *SharkieFilesystem) ReadFullFd(fd FileDescriptor) ([]byte, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return nil, errors.New("invalid file descriptor")
	}

	return shFile.ReadFull()
}

func (shFs *SharkieFilesystem) Write(path string, data []byte) (int, error) {
	fd, err := shFs.Open(path, O_CREAT|O_WRONLY|O_TRUNC, 0777)
	if err != nil {
		return 0, err
	}
	defer shFs.Close(fd)
	return shFs.WriteFd(fd, data)
}

func (shFs *SharkieFilesystem) WriteFd(fd FileDescriptor, data []byte) (int, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Write(data)
}

func (shFs *SharkieFilesystem) Pwrite(path string, data []byte, offset int64) (int, error) {
	fd, err := shFs.Open(path, O_CREAT|O_WRONLY|O_TRUNC, 0777)
	if err != nil {
		return 0, err
	}
	defer shFs.Close(fd)
	return shFs.PwriteFd(fd, data, offset)
}

func (shFs *SharkieFilesystem) PwriteFd(fd FileDescriptor, data []byte, offset int64) (int, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return 0, errors.New("invalid file descriptor")
	}

	return shFile.Pwrite(data, offset)
}

func (shFs *SharkieFilesystem) Truncate(path string, length int64) error {
	fd, err := shFs.Open(path, O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer shFs.Close(fd)
	return shFs.TruncateFd(fd, length)
}

func (shFs *SharkieFilesystem) TruncateFd(fd FileDescriptor, length int64) error {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return errors.New("invalid file descriptor")
	}

	return shFile.Truncate(length)
}

func (shFs *SharkieFilesystem) Ioctl(path string, request uint64, argPtr uintptr) error {
	fd, err := shFs.Open(path, O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer shFs.Close(fd)
	return shFs.IoctlFd(fd, request, argPtr)
}

func (shFs *SharkieFilesystem) IoctlFd(fd FileDescriptor, request uint64, argPtr uintptr) error {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return errors.New("invalid file descriptor")
	}

	return shFile.Ioctl(request, argPtr)
}

func (shFs *SharkieFilesystem) Stat(path string) (*FileStat, error) {
	fd, err := shFs.Open(path, O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer shFs.Close(fd)
	return shFs.StatFd(fd)
}

func (shFs *SharkieFilesystem) StatFd(fd FileDescriptor) (*FileStat, error) {
	shFs.Lock.Lock()
	shFile, ok := shFs.Descriptors[fd]
	shFs.Lock.Unlock()
	if !ok {
		return nil, errors.New("invalid file descriptor")
	}

	return shFile.Stat()
}

func (shFs *SharkieFilesystem) Close(fd FileDescriptor) error {
	shFs.Lock.Lock()
	defer shFs.Lock.Unlock()
	shFile, ok := shFs.Descriptors[fd]
	if !ok {
		return errors.New("invalid file descriptor")
	}

	if err := shFile.Close(); err != nil {
		return err
	}
	delete(shFs.Descriptors, fd)

	return nil
}

// MkdirAll creates a new directory in VFS.
func (shFs *SharkieFilesystem) MkdirAll(path string) error {
	return shFs.Fs.MkdirAll(path, 0777)
}

// Mount mounts a directory in VFS.
func (shFs *SharkieFilesystem) Mount(path string, hostPath string, readOnly bool) error {
	return shFs.Fs.Mount(path, hostPath, readOnly)
}

// Unmount unmounts a directory in VFS.
func (shFs *SharkieFilesystem) Unmount(path string) error {
	return shFs.Fs.Unmount(path)
}

// Delete removes a file from the VFS. This does not close active file descriptors pointing to it.
func (shFs *SharkieFilesystem) Delete(path string) error {
	return shFs.Fs.Remove(path)
}

func SetupFilesystem() {
	GlobalFilesystem = NewFilesystem()
}
