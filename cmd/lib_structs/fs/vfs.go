package fs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FS represents a POSIX-style in-memory file system tree.
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
	hostPath string
	ReadOnly bool
	size     int64
	children map[string]*Node
	mu       sync.RWMutex
}

// NewFS creates a new in-memory file system tree.
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
		curr.mu.Lock()
		child, ok := curr.children[part]
		if !ok && curr.hostPath != "" {
			targetHostPath := filepath.Join(curr.hostPath, part)
			info, err := os.Stat(targetHostPath)
			if err == nil {
				child = &Node{
					name:     part,
					isDir:    info.IsDir(),
					mode:     info.Mode(),
					modTime:  info.ModTime(),
					hostPath: targetHostPath,
					ReadOnly: curr.ReadOnly,
					size:     info.Size(),
					children: make(map[string]*Node),
				}
				curr.children[part] = child
				ok = true
			}
		}
		curr.mu.Unlock()
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

// GetHostPath returns the underlying host path for a given virtual path, if it exists.
func (fsys *FS) GetHostPath(path string) (string, error) {
	fsys.mu.RLock()
	defer fsys.mu.RUnlock()

	dirPath := path
	baseName := ""
	if idx := strings.LastIndex(path, "/"); idx >= 0 && idx < len(path)-1 {
		dirPath = path[:idx]
		baseName = path[idx+1:]
	}

	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return "", err
	}
	if baseName == "" {
		if dir.hostPath == "" {
			return "", errors.New("node has no host path")
		}
		return dir.hostPath, nil
	}

	dir.mu.RLock()
	defer dir.mu.RUnlock()
	node, ok := dir.children[baseName]
	if !ok {
		// Try to construct it dynamically if directory has a hostPath.
		if dir.hostPath != "" {
			targetHostPath := filepath.Join(dir.hostPath, baseName)
			if _, err = os.Stat(targetHostPath); err == nil {
				return targetHostPath, nil
			}
		}
		return "", fs.ErrNotExist
	}
	if node.hostPath == "" {
		return "", errors.New("node has no host path")
	}

	return node.hostPath, nil
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

// GetOrCreateNode retrieves a node or creates it if it doesn't exist and create is true.
func (fsys *FS) GetOrCreateNode(name string, create bool, excl bool, perm os.FileMode) (*Node, error) {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	// Split path into directory and target filename.
	name = strings.TrimLeft(name, "/")
	dirPath, baseName := ".", name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		dirPath = name[:idx]
		baseName = name[idx+1:]
	}

	// Find the parent directory.
	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return nil, err
	}
	dir.mu.Lock()
	defer dir.mu.Unlock()

	node, exists := dir.children[baseName]
	var targetHostPath string
	if dir.hostPath != "" {
		targetHostPath = filepath.Join(dir.hostPath, baseName)
	}
	if !exists && targetHostPath != "" {
		if info, err := os.Stat(targetHostPath); err == nil {
			node = &Node{
				name:     baseName,
				isDir:    info.IsDir(),
				mode:     info.Mode(),
				modTime:  info.ModTime(),
				hostPath: targetHostPath,
				ReadOnly: dir.ReadOnly,
				size:     info.Size(),
				children: make(map[string]*Node),
			}
			dir.children[baseName] = node
			exists = true
		}
	}
	if !exists {
		if !create {
			return nil, fs.ErrNotExist
		}
		if dir.ReadOnly {
			return nil, fs.ErrPermission
		}
		if targetHostPath != "" {
			f, err := os.OpenFile(targetHostPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
			if err != nil {
				return nil, err
			}
			f.Close()
			node = &Node{
				name:     baseName,
				isDir:    false,
				mode:     perm,
				modTime:  time.Now(),
				hostPath: targetHostPath,
				ReadOnly: false,
				size:     0,
				children: make(map[string]*Node),
			}
		} else {
			node = &Node{
				name:    baseName,
				isDir:   false,
				mode:    perm,
				modTime: time.Now(),
				data:    make([]byte, 0),
			}
		}
		dir.children[baseName] = node
	} else if create && excl {
		return nil, fs.ErrExist
	}

	return node, nil
}

// Remove deletes the named file or (empty) directory from the virtual filesystem.
func (fsys *FS) Remove(name string) error {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	name = strings.TrimLeft(name, "/")
	if name == "" || name == "." {
		return errors.New("cannot remove root directory")
	}

	// Split path into directory and target filename.
	dirPath, baseName := ".", name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		dirPath = name[:idx]
		baseName = name[idx+1:]
	}

	// Find the parent directory.
	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return err
	}
	if dir.ReadOnly {
		return fs.ErrPermission
	}
	dir.mu.Lock()
	defer dir.mu.Unlock()
	node, exists := dir.children[baseName]
	if !exists {
		return fs.ErrNotExist
	}

	// Optional POSIX check: if it's a directory, ensure it's empty before removal.
	if node.isDir && len(node.children) > 0 {
		return errors.New("directory not empty")
	}
	// TODO: remove if mounted.

	// Remove from the parent's map.
	delete(dir.children, baseName)
	dir.modTime = time.Now()

	return nil
}

// MapHostFile creates a node that points to a file on the host OS.
func (fsys *FS) MapHostFile(name string, hostPath string, size int64, perm os.FileMode) error {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()

	// Split path into directory and target filename.
	name = strings.TrimLeft(name, "/")
	dirPath, baseName := ".", name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		dirPath = name[:idx]
		baseName = name[idx+1:]
	}

	// Find the parent directory.
	dir, err := fsys.resolveDir(dirPath)
	if err != nil {
		return err
	}
	dir.mu.Lock()
	defer dir.mu.Unlock()
	if _, exists := dir.children[baseName]; exists {
		return fs.ErrExist
	}

	dir.children[baseName] = &Node{
		name:     baseName,
		isDir:    false,
		mode:     perm,
		modTime:  time.Now(),
		hostPath: hostPath,
		size:     size,
	}

	return nil
}

// Mount maps a host OS folder into the virtual filesystem.
func (fsys *FS) Mount(path string, hostPath string, readOnly bool) error {
	if err := fsys.MkdirAll(path, 0777); err != nil {
		return err
	}

	// Find the parent directory.
	node, err := fsys.resolveDir(path)
	if err != nil {
		return err
	}
	node.mu.Lock()
	defer node.mu.Unlock()

	node.hostPath = hostPath
	node.ReadOnly = readOnly
	node.children = make(map[string]*Node)

	return nil
}

// Unmount unmaps a host OS folder from the virtual filesystem.
func (fsys *FS) Unmount(path string) error {
	// Find the parent directory.
	node, err := fsys.resolveDir(path)
	if err != nil {
		return err
	}
	node.mu.Lock()
	defer node.mu.Unlock()

	node.hostPath = ""
	node.ReadOnly = false
	node.children = make(map[string]*Node)

	return nil
}
