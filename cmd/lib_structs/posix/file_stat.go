package posix

import (
	"io/fs"
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
)

const (
	S_IFMT   = 0170000 // Type mask
	S_IFIFO  = 0010000 // Named pipe (fifo)
	S_IFCHR  = 0020000 // Character special
	S_IFDIR  = 0040000 // Directory
	S_IFBLK  = 0060000 // Block special
	S_IFREG  = 0100000 // Regular
	S_IFLNK  = 0120000 // Symbolic link
	S_IFSOCK = 0140000 // Socket
)

type FileStat struct {
	Device                uint32    // st_dev
	Inodes                uint32    // st_ino
	Mode                  uint16    // st_mode
	HardLinkCount         uint16    // st_nlink
	OwnerUser             uint32    // st_uid
	OwnerGroup            uint32    // st_gid
	SpecialDevice         uint32    // st_rdev
	AccessTime            Timestamp // st_atim
	ModifyTime            Timestamp // st_mtim
	ChangeStatusTime      Timestamp // st_ctim
	Size                  int64     // st_size
	Blocks                int64     // st_blocks
	BlockSize             uint32    // st_blksize
	Flags                 uint32    // st_flags
	GenerationNumber      uint32    // st_gen
	ImplementationDetails int32     // st_lspare
	CreateTime            Timestamp // st_birthtim
}

const FileStatSize = unsafe.Sizeof(FileStat{})

type DirectoryEntry struct {
	FileNumber   uint32    // d_fileno
	RecordLength uint16    // d_reclen
	Type         uint8     // d_type
	NameLength   uint8     // d_namlen
	Name         [256]byte // d_name
}

// fileno(4) + reclen(2) + type(1) + namlen(1) = 8 bytes.
const DirectoryEntryHeaderSize = uintptr(8)
const DirectoryEntrySize = unsafe.Sizeof(DirectoryEntry{})

// HashFilenameToInode creates a stable, non-zero 32-bit ID using FNV-1a.
func HashFilenameToInode(name string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(name); i++ {
		hash ^= uint32(name[i])
		hash *= 16777619
	}
	if hash == 0 {
		return 1
	}

	return hash
}

// GoToPosixMode translates a Go fs.FileMode to a POSIX st_mode.
func GoToPosixMode(mode fs.FileMode) uint16 {
	posixMode := uint16(mode & fs.ModePerm)
	switch {
	case mode.IsDir():
		posixMode |= S_IFDIR
	case mode&fs.ModeSymlink != 0:
		posixMode |= S_IFLNK
	case mode&fs.ModeDevice != 0:
		if mode&fs.ModeCharDevice != 0 {
			posixMode |= S_IFCHR
		} else {
			posixMode |= S_IFBLK
		}
	case mode&fs.ModeSocket != 0:
		posixMode |= S_IFSOCK
	case mode&fs.ModeNamedPipe != 0:
		posixMode |= S_IFIFO
	default:
		posixMode |= S_IFREG
	}

	return posixMode
}
