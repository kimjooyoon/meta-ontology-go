package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func requireEmpty(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("destination is not empty: %s", path)
	}
	return nil
}

func moveChildren(source, destination string, preserveGit bool) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	for _, item := range entries {
		if preserveGit && item.Name() == ".git" {
			continue
		}
		from := filepath.Join(source, item.Name())
		to := filepath.Join(destination, item.Name())
		if _, err := os.Lstat(to); !os.IsNotExist(err) {
			return fmt.Errorf("activation collision: %s", to)
		}
		if err := os.Rename(from, to); err != nil {
			return err
		}
	}
	return nil
}

func leafCount(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, item fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != root && filepath.Dir(path) == root && item.Name() == ".git" {
			if item.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if path != root && !item.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}
