package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
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

	// Directory properties.
	DirEntries []fs.FileInfo
	DirIndex   int
	DirLoaded  bool

	// Device file properties.
	Device   PosixFile
	IsDevice bool

	mu sync.Mutex
}

func (shFile *SharkieFile) Read(data []byte) (int, error) {
	if shFile.IsDevice {
		return shFile.Device.Read(data)
	}
	if shFile.Node.isDir {
		return 0, errors.New("is a directory")
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
	if shFile.Node.isDir {
		return 0, errors.New("is a directory")
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
	if shFile.Node.isDir {
		return []byte{}, errors.New("is a directory")
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
	if shFile.Node.isDir {
		return 0, errors.New("is a directory")
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Check if file was opened with write permissions.
	if shFile.Flags&os.O_RDWR == 0 && shFile.Flags&os.O_WRONLY == 0 {
		return 0, errors.New("file not opened for writing")
	}
	if shFile.Flags&os.O_APPEND != 0 {
		shFile.Offset = shFile.Node.GetSize()
	}
	n, err := shFile.Node.WriteAt(data, shFile.Offset)
	if err == nil {
		shFile.Offset += int64(n)
	}

	return n, err
}

func (shFile *SharkieFile) Pwrite(data []byte, offset int64) (int, error) {
	if shFile.IsDevice {
		return 0, errors.New("illegal seek")
	}
	if shFile.Node.isDir {
		return 0, errors.New("is a directory")
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

	// Directory seek.
	if shFile.Node.isDir {
		if whence == io.SeekStart && offset == 0 {
			shFile.DirIndex = 0
			shFile.DirLoaded = false
			shFile.Offset = 0
			return 0, nil
		}
		if whence == io.SeekCurrent && offset == 0 {
			return shFile.Offset, nil
		}
		return 0, errors.New("invalid whence")
	}

	// File seek.
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
	if shFile.Node.isDir {
		return errors.New("is a directory")
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

func (shFile *SharkieFile) Close() error {
	if shFile.IsDevice {
		return shFile.Device.Close()
	}

	return nil
}

// Stat translates the generic underlying data into the PS4-specific FileStat struct.
func (shFile *SharkieFile) Stat() (*FileStat, error) {
	if shFile.IsDevice {
		return &FileStat{
			Mode:          S_IFCHR | 0666,
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
		Mode:             GoToPosixMode(nodeStat.Mode()),
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

// Getdents reads directory entries into a buffer.
func (shFile *SharkieFile) Getdents(nbytes uint64) ([]byte, error) {
	if shFile.IsDevice {
		return nil, errors.New("not a directory")
	}
	if !shFile.Node.isDir {
		return nil, errors.New("not a directory")
	}
	shFile.mu.Lock()
	defer shFile.mu.Unlock()

	// Lazy load directory entries.
	if !shFile.DirLoaded {
		entries, err := shFile.Node.ReadDir()
		if err != nil {
			return nil, err
		}
		selfInfo, _ := shFile.Node.Stat()
		shFile.DirEntries = append([]fs.FileInfo{
			&fileInfo{name: ".", mode: fs.ModeDir | 0777, modTime: selfInfo.ModTime()},
			&fileInfo{name: "..", mode: fs.ModeDir | 0777, modTime: selfInfo.ModTime()},
		}, entries...)
		shFile.DirLoaded = true
	}
	if shFile.DirIndex >= len(shFile.DirEntries) {
		return []byte{}, nil
	}

	// Populate directory entries.
	var buffer []byte
	for shFile.DirIndex < len(shFile.DirEntries) {
		entry := shFile.DirEntries[shFile.DirIndex]
		nameLength := uint64(len(entry.Name()))
		if nameLength > 255 {
			nameLength = 255
		}

		// Entry aligned to 4-bytes (+1 for NULL terminator).
		recordLength := ((uint64(DirectoryEntryHeaderSize) + nameLength + 1) + 3) &^ 3
		if uint64(len(buffer))+recordLength > nbytes {
			break // Next entry won't fit in the requested chunk.
		}

		// Construct directory entry.
		directoryEntry := DirectoryEntry{
			FileNumber:   HashFilenameToInode(entry.Name()),
			RecordLength: uint16(recordLength),
			Type:         uint8(GoToPosixMode(entry.Mode()) >> 12),
			NameLength:   uint8(nameLength),
		}
		copy(directoryEntry.Name[:], entry.Name()[:nameLength])

		// Append to buffer.
		directoryEntryBytes := unsafe.Slice((*byte)(unsafe.Pointer(&directoryEntry)), recordLength)
		buffer = append(buffer, directoryEntryBytes...)
		shFile.DirIndex++
	}
	shFile.Offset += int64(len(buffer))

	return buffer, nil
}
