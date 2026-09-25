package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func restoreEntries(model manifest, physical, destination string) (int, error) {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return 0, err
	}
	restored := 0
	for _, item := range model.Entries {
		backing, err := resolvePath(physical, item.Backing)
		if err != nil {
			return restored, err
		}
		data, err := backingData(backing)
		if err != nil {
			return restored, err
		}
		if contentDigest(data) != item.ContentSHA || objectDigest(item, data) != item.ObjectSHA {
			return restored, fmt.Errorf("projection digest mismatch: %s", item.Logical)
		}
		target, err := resolvePath(destination, item.Logical)
		if err != nil {
			return restored, err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return restored, err
		}
		if item.Kind == "symlink" {
			err = os.Symlink(string(data), target)
		} else {
			err = os.WriteFile(target, data, os.FileMode(item.Mode))
		}
		if err != nil {
			return restored, err
		}
		restored++
	}
	return restored, nil
}

func backingData(name string) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(name)
		return []byte(target), err
	}
	return os.ReadFile(name)
}

func contentDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func objectDigest(item entry, data []byte) string {
	class := "blob"
	if item.Kind == "symlink" {
		class = "symlink"
	} else if item.Language == "go" || item.Language == "gooo" {
		class = "source"
	}
	payload := append(append([]byte(class), 0), data...)
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
