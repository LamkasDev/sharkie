package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
)

// FS represents a writable, POSIX-style in-memory file system.
type FS struct {
	mu   sync.RWMutex
	root *Node
}

// Node represents a file or directory in memory.
type Node struct {
	name     string
	isDir    bool
	mode     fs.FileMode
	modTime  time.Time
	data     []byte
	children map[string]*Node
	mu       sync.RWMutex
}

// NewFS creates a new in-memory file system.
func NewFS() *FS {
	return &FS{
		root: &Node{
			name:     "/",
			isDir:    true,
			mode:     fs.ModeDir | 0777,
			modTime:  time.Now(),
			children: make(map[string]*Node),
		},
	}
}

// resolveDir traverses the path to find the target directory node.
func (fsys *FS) resolveDir(path string) (*Node, error) {
	path = strings.TrimLeft(path, "/")
	if path == "" || path == "." {
		return fsys.root, nil
	}
	parts := strings.Split(path, "/")
	curr := fsys.root

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		curr.mu.RLock()
		child, ok := curr.children[part]
		curr.mu.RUnlock()
		if !ok {
			return nil, fs.ErrNotExist
		}
		if !child.isDir {
			return nil, fs.ErrInvalid
		}
		curr = child
	}
	return curr, nil
}

// MkdirAll creates a directory and all its parents.
func (fsys *FS) MkdirAll(path string, perm os.FileMode) error {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	path = strings.TrimLeft(path, "/")
	parts := strings.Split(path, "/")
	curr := fsys.root

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		curr.mu.Lock()
		child, ok := curr.children[part]
		if !ok {
			child = &Node{
				name:     part,
				isDir:    true,
				mode:     fs.ModeDir | perm,
				modTime:  time.Now(),
				children: make(map[string]*Node),
			}
			curr.children[part] = child
		} else if !child.isDir {
			curr.mu.Unlock()
			return fs.ErrExist
		}
		curr.mu.Unlock()
		curr = child
	}
	return nil
}

// OpenFile opens a file with standard OS flags (os.O_RDWR, os.O_CREATE, etc.).
func (fsys *FS) OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	name = strings.TrimLeft(name, "/")
	dirPath, baseName := ".", name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		dirPath = name[:idx]
		baseName = name[idx+1:]
	}

	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return nil, err
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	node, exists := dir.children[baseName]
	if !exists {
		if flag&os.O_CREATE == 0 {
			return nil, fs.ErrNotExist
		}
		node = &Node{
			name:    baseName,
			isDir:   false,
			mode:    perm,
			modTime: time.Now(),
			data:    make([]byte, 0),
		}
		dir.children[baseName] = node
	} else {
		if flag&os.O_CREATE != 0 && flag&os.O_EXCL != 0 {
			return nil, fs.ErrExist
		}
		if node.isDir {
			return nil, errors.New("is a directory")
		}
		if flag&os.O_TRUNC != 0 {
			node.mu.Lock()
			node.data = node.data[:0]
			node.modTime = time.Now()
			node.mu.Unlock()
		}
	}

	f := &File{node: node, flag: flag}
	if flag&os.O_APPEND != 0 {
		f.offset = int64(len(node.data))
	}
	return f, nil
}

// Open is a convenience wrapper for io/fs compatibility.
func (fsys *FS) Open(name string) (fs.File, error) {
	return fsys.OpenFile(name, os.O_RDONLY, 0)
}

// WriteFile is a helper utility for easy data dumping.
func (fsys *FS) WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := fsys.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// ReadFile is a helper utility for easy data reading.
func (fsys *FS) ReadFile(name string) ([]byte, error) {
	f, err := fsys.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// Remove deletes the named file or (empty) directory from the virtual filesystem.
func (fsys *FS) Remove(name string) error {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	name = strings.TrimLeft(name, "/")
	if name == "" || name == "." {
		return errors.New("cannot remove root directory")
	}

	// Split path into directory and target filename
	dirPath, baseName := ".", name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		dirPath = name[:idx]
		baseName = name[idx+1:]
	}

	// Find the parent directory
	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return err
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	// Check if the target exists
	node, exists := dir.children[baseName]
	if !exists {
		return fs.ErrNotExist
	}

	// Optional POSIX check: if it's a directory, ensure it's empty before removal
	if node.isDir && len(node.children) > 0 {
		return errors.New("directory not empty")
	}

	// Remove from the parent's map
	delete(dir.children, baseName)
	dir.modTime = time.Now()

	return nil
}

// File represents an open file descriptor with an independent seek offset.
type File struct {
	node   *Node
	offset int64
	flag   int
	closed bool
	mu     sync.Mutex
}

func (f *File) Read(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}

	f.node.mu.RLock()
	defer f.node.mu.RUnlock()

	if f.offset >= int64(len(f.node.data)) {
		return 0, io.EOF
	}

	n := copy(b, f.node.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *File) Write(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	// Check if opened for writing
	if f.flag&os.O_RDWR == 0 && f.flag&os.O_WRONLY == 0 {
		return 0, errors.New("file not opened for writing")
	}

	f.node.mu.Lock()
	defer f.node.mu.Unlock()

	end := f.offset + int64(len(b))
	if end > int64(len(f.node.data)) {
		newData := make([]byte, end)
		copy(newData, f.node.data)
		f.node.data = newData
	}

	n := copy(f.node.data[f.offset:], b)
	f.offset += int64(n)
	f.node.modTime = time.Now()
	return n, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}

	f.node.mu.RLock()
	defer f.node.mu.RUnlock()

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = f.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(f.node.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if newOffset < 0 {
		return 0, errors.New("negative seek offset")
	}
	f.offset = newOffset
	return newOffset, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	if f.closed {
		return nil, fs.ErrClosed
	}
	f.node.mu.RLock()
	defer f.node.mu.RUnlock()
	return &fileInfo{
		name:    f.node.name,
		size:    int64(len(f.node.data)),
		modTime: f.node.modTime,
		mode:    f.node.mode,
	}, nil
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *File) Truncate(size int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return fs.ErrClosed
	}

	f.node.mu.Lock()
	defer f.node.mu.Unlock()

	if size < 0 {
		return errors.New("negative size")
	}

	currentSize := int64(len(f.node.data))
	if size == currentSize {
		return nil
	}

	if size < currentSize {
		// Shrink
		f.node.data = f.node.data[:size]
	} else {
		// Expand with zeros
		expansion := make([]byte, size-currentSize)
		f.node.data = append(f.node.data, expansion...)
	}

	f.node.modTime = time.Now()
	return nil
}

func (f *File) Ioctl(request uint64, argPtr uintptr) error {
	return errors.New("not implemented")
}

// fileInfo implements fs.FileInfo
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
