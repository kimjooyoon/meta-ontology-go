package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func applyCutover(settings cutoverConfig) error {
	if err := requireEmpty(settings.backup); err != nil {
		return err
	}
	if err := moveRoot(settings.root, settings.backup, true); err != nil {
		return err
	}
	return moveRoot(settings.physical, settings.root, false)
}

func requireEmpty(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("cutover destination is not empty: %s", path)
	}
	return nil
}

func moveRoot(source, destination string, preserveGit bool) error {
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
			return fmt.Errorf("cutover collision: %s", to)
		}
		if err := os.Rename(from, to); err != nil {
			return err
		}
	}
	return nil
}
