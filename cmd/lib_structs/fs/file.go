package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
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

	// Regular file properties.
	Node   *Node
	Offset int64
	Flags  int

	// Device file properties.
	Device   PosixFile
	IsDevice bool

	mu sync.Mutex
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
	if shFile.IsDevice {
		return shFile.Device.Read(data)
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with read permissions.
	if shFile.Flags&os.O_WRONLY != 0 {
		return 0, errors.New("file not opened for reading")
	}
	n, err := shFile.Node.ReadAt(data, shFile.Offset)
	shFile.Offset += int64(n)

	return n, err
}

func (shFile *SharkieFile) Pread(data []byte, offset int64) (int, error) {
	if shFile.IsDevice {
		return 0, errors.New("illegal seek")
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with read permissions.
	if shFile.Flags&os.O_WRONLY != 0 {
		return 0, errors.New("file not opened for reading")
	}

	return shFile.Node.ReadAt(data, offset)
}

func (shFile *SharkieFile) ReadFull() ([]byte, error) {
	if shFile.IsDevice {
		return []byte{}, nil
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Read from offset 0 to get the entire file.
	size := shFile.Node.GetSize()
	buffer := make([]byte, size)
	_, err := shFile.Node.ReadAt(buffer, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buffer, nil
}

func (shFile *SharkieFile) Write(data []byte) (int, error) {
	if shFile.IsDevice {
		return shFile.Device.Write(data)
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with write permissions.
	if shFile.Flags&os.O_RDWR == 0 && shFile.Flags&os.O_WRONLY == 0 {
		return 0, errors.New("file not opened for writing")
	}
	n, err := shFile.Node.WriteAt(data, shFile.Offset)
	shFile.Offset += int64(n)

	return n, err
}

func (shFile *SharkieFile) Pwrite(data []byte, offset int64) (int, error) {
	if shFile.IsDevice {
		return 0, errors.New("illegal seek")
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with write permissions.
	if shFile.Flags&os.O_RDWR == 0 && shFile.Flags&os.O_WRONLY == 0 {
		return 0, errors.New("file not opened for writing")
	}

	return shFile.Node.WriteAt(data, offset)
}

func (shFile *SharkieFile) Seek(offset int64, whence int) (int64, error) {
	if shFile.IsDevice {
		return shFile.Device.Seek(offset, whence)
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = shFile.Offset + offset
	case io.SeekEnd:
		// Quickly check the exact size of the node safely.
		size := shFile.Node.GetSize()
		newOffset = size + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if newOffset < 0 {
		return 0, errors.New("negative seek offset")
	}
	shFile.Offset = newOffset

	return newOffset, nil
}

func (shFile *SharkieFile) Truncate(length int64) error {
	if shFile.IsDevice {
		return shFile.Device.Truncate(length)
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with write permissions.
	if shFile.Flags&os.O_RDWR == 0 && shFile.Flags&os.O_WRONLY == 0 {
		return errors.New("file not opened for writing")
	}

	return shFile.Node.Truncate(length)
}

func (shFile *SharkieFile) Ioctl(request uint64, argPtr uintptr) error {
	if shFile.IsDevice {
		return shFile.Device.Ioctl(request, argPtr)
	}

	return errors.New("inappropriate ioctl for device")
}

// Stat translates the generic underlying data into the PS4-specific FileStat struct.
func (shFile *SharkieFile) Stat() (*FileStat, error) {
	if shFile.IsDevice {
		return &FileStat{
			Mode:          020666,
			BlockSize:     FileBlockSize,
			HardLinkCount: 1,
		}, nil
	}

	nodeStat, err := shFile.Node.Stat()
	if err != nil {
		return nil, err
	}
	modTime := nodeStat.ModTime()
	timestamp := Timestamp{
		Seconds:     modTime.Unix(),
		Nanoseconds: int64(modTime.Nanosecond()),
	}
	stat := &FileStat{
		Mode:             uint16(nodeStat.Mode()),
		Size:             nodeStat.Size(),
		BlockSize:        FileBlockSize,
		HardLinkCount:    1,
		Blocks:           (nodeStat.Size() + 511) / 512,
		AccessTime:       timestamp,
		ModifyTime:       timestamp,
		ChangeStatusTime: timestamp,
		CreateTime:       timestamp,
	}

	return stat, nil
}

func (shFile *SharkieFile) Close() error {
	if shFile.IsDevice {
		return shFile.Device.Close()
	}

	return nil
}
