package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func materialize(stored, output string, model manifest) (int, error) {
	if err := os.MkdirAll(output, 0o755); err != nil {
		return 0, err
	}
	for _, entry := range model.Entries {
		if err := safeLogical(entry.Logical); err != nil {
			return 0, err
		}
		if err := safeLogical(entry.Backing); err != nil {
			return 0, err
		}
		data, err := os.ReadFile(filepath.Join(stored, filepath.FromSlash(entry.Backing)))
		if err != nil {
			return 0, err
		}
		class := objectClass(entry.Kind, entry.Language)
		if contentHash(data) != entry.ContentSHA || objectHash(class, data) != entry.ObjectSHA {
			return 0, fmt.Errorf("object evidence mismatch for %s", entry.Logical)
		}
		target := filepath.Join(output, filepath.FromSlash(entry.Logical))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return 0, err
		}
		if entry.Kind == "symlink" {
			err = os.Symlink(string(data), target)
		} else {
			err = os.WriteFile(target, data, os.FileMode(entry.Mode))
		}
		if err != nil {
			return 0, err
		}
	}
	return compareMaterialized(output, model)
}

func compareMaterialized(root string, model manifest) (int, error) {
	loss := 0
	for _, entry := range model.Entries {
		name := filepath.Join(root, filepath.FromSlash(entry.Logical))
		info, err := os.Lstat(name)
		if err != nil {
			loss++
			continue
		}
		var data []byte
		if entry.Kind == "symlink" {
			target, readErr := os.Readlink(name)
			if readErr != nil || info.Mode()&os.ModeSymlink == 0 {
				loss++
				continue
			}
			data = []byte(target)
		} else {
			data, err = os.ReadFile(name)
			if err != nil || !info.Mode().IsRegular() || uint32(info.Mode().Perm()) != entry.Mode {
				loss++
				continue
			}
		}
		if contentHash(data) != entry.ContentSHA {
			loss++
		}
	}
	return loss, nil
}
