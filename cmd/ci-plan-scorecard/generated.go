package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func equalGeneratedGo(first, second string) (bool, error) {
	left, err := generatedGoFiles(first)
	if err != nil {
		return false, err
	}
	right, err := generatedGoFiles(second)
	if err != nil {
		return false, err
	}
	if len(left) == 0 || len(left) != len(right) {
		return false, nil
	}
	for path, raw := range left {
		if !bytes.Equal(raw, right[path]) {
			return false, nil
		}
	}
	return true, nil
}

func generatedGoFiles(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	paths := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		files[filepath.ToSlash(relative)] = raw
	}
	return files, nil
}
