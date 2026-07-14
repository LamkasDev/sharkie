package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"time"
)

// ensureHostFileLoaded checks if the node points to a host file and loads it into memory if needed.
func (node *Node) ensureHostFileLoaded() error {
	if node.hostPath != "" && node.data == nil {
		data, err := os.ReadFile(node.hostPath)
		if err != nil {
			return err
		}
		node.data = data
		node.size = int64(len(data))
	}

	return nil
}

// ReadAt reads len(b) bytes from the Node starting at byte offset off.
// It implements io.ReaderAt.
func (node *Node) ReadAt(b []byte, off int64) (int, error) {
	node.mu.RLock()
	defer node.mu.RUnlock()
	if node.isDir {
		return 0, errors.New("is a directory")
	}
	if node.hostPath != "" {
		f, err := os.Open(node.hostPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		return f.ReadAt(b, off)
	}
	if off < 0 {
		return 0, errors.New("negative offset")
	}
	if off >= int64(len(node.data)) {
		return 0, io.EOF
	}

	n := copy(b, node.data[off:])
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// WriteAt writes len(b) bytes to the Node starting at byte offset off.
// It implements io.WriterAt.
func (node *Node) WriteAt(b []byte, off int64) (int, error) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.ReadOnly {
		return 0, fs.ErrPermission
	}
	if node.isDir {
		return 0, errors.New("is a directory")
	}
	if node.hostPath != "" {
		f, err := os.OpenFile(node.hostPath, os.O_WRONLY, 0)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		n, err := f.WriteAt(b, off)
		if err == nil {
			node.modTime = time.Now()
		}
		return n, err
	}
	if off < 0 {
		return 0, errors.New("negative offset")
	}

	end := off + int64(len(b))
	if end > int64(len(node.data)) {
		newData := make([]byte, end)
		copy(newData, node.data)
		node.data = newData
		node.size = end
	}

	n := copy(node.data[off:], b)
	node.modTime = time.Now()

	return n, nil
}

// Truncate changes the size of the file.
func (node *Node) Truncate(size int64) error {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.ReadOnly {
		return fs.ErrPermission
	}
	if node.isDir {
		return errors.New("is a directory")
	}
	if size < 0 {
		return errors.New("negative size")
	}
	if node.hostPath != "" {
		err := os.Truncate(node.hostPath, size)
		if err == nil {
			node.size = size
			node.modTime = time.Now()
		}
		return err
	}

	currentSize := int64(len(node.data))
	if size == currentSize {
		return nil
	}
	if size < currentSize {
		node.data = node.data[:size]
	} else {
		expansion := make([]byte, size-currentSize)
		node.data = append(node.data, expansion...)
	}

	node.size = size
	node.modTime = time.Now()

	return nil
}

// GetSize returns size of the file.
func (node *Node) GetSize() int64 {
	if node.hostPath != "" {
		if info, err := os.Stat(node.hostPath); err == nil {
			return info.Size()
		}
	}
	if node.data != nil {
		return int64(len(node.data))
	}
	return node.size
}

// Stat returns the fs.FileInfo for the node.
func (node *Node) Stat() (fs.FileInfo, error) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	return &fileInfo{
		name:    node.name,
		size:    node.GetSize(),
		modTime: node.modTime,
		mode:    node.mode,
	}, nil
}

// fileInfo implements fs.FileInfo.
type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	mode    fs.FileMode
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi *fileInfo) Sys() interface{}   { return nil }
