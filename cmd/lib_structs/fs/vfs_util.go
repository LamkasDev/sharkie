package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// resolveHostPath performs a case-insensitive search on the physical host OS.
func resolveHostPath(parentPath, name string) (string, os.FileInfo, error) {
	exactPath := filepath.Join(parentPath, name)
	info, err := os.Stat(exactPath)
	if err == nil {
		return exactPath, info, nil
	}
	if !os.IsNotExist(err) {
		return "", nil, err
	}

	// Read directory and do a case-insensitive search.
	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return "", nil, err
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			actualPath := filepath.Join(parentPath, entry.Name())
			info, err = os.Stat(actualPath)
			if err != nil {
				return "", nil, err
			}
			return actualPath, info, nil
		}
	}

	return "", nil, fs.ErrNotExist
}

// findVirtualChild performs a case-insensitive search in a virtual Node's children map.
func findVirtualChild(dir *Node, name string) (*Node, string) {
	if child, ok := dir.children[name]; ok {
		return child, name
	}
	for k, v := range dir.children {
		if strings.EqualFold(k, name) {
			return v, k
		}
	}

	return nil, ""
}
