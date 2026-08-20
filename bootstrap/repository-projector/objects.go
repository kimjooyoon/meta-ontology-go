package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func readTracked(root string, paths []string) ([]trackedFile, error) {
	files := make([]trackedFile, 0, len(paths))
	for _, logical := range paths {
		if err := safeLogical(logical); err != nil {
			return nil, err
		}
		name := filepath.Join(root, filepath.FromSlash(logical))
		info, err := os.Lstat(name)
		if err != nil {
			return nil, err
		}
		file := trackedFile{logical: logical, mode: uint32(info.Mode().Perm())}
		file.language = languageFor(logical)
		switch {
		case info.Mode().IsRegular():
			file.kind = "file"
			file.data, err = os.ReadFile(name)
		case info.Mode()&os.ModeSymlink != 0:
			file.kind = "symlink"
			var target string
			target, err = os.Readlink(name)
			file.data = []byte(target)
		default:
			return nil, fmt.Errorf("unsupported tracked kind: %s", logical)
		}
		if err != nil {
			return nil, err
		}
		file.lines = lineCount(file.data)
		file.objectSHA = objectHash(objectClass(file.kind, file.language), file.data)
		files = append(files, file)
	}
	return files, nil
}

func objectClass(kind, language string) string {
	if kind == "symlink" {
		return "symlink"
	}
	if language == "go" || language == "gooo" {
		return "source"
	}
	return "blob"
}

func objectHash(class string, data []byte) string {
	digest := sha256.New()
	digest.Write([]byte(class))
	digest.Write([]byte{0})
	digest.Write(data)
	return hex.EncodeToString(digest.Sum(nil))
}

func contentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
