package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeAtomic(path string, payload []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	temporary, err := os.CreateTemp(
		directory,
		"."+filepath.Base(path)+".*",
	)
	if err != nil {
		return fmt.Errorf("create temporary report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		temporary.Close()
		os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(0o644); err != nil {
		return fmt.Errorf("set report mode: %w", err)
	}
	if _, err := temporary.Write(payload); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace report: %w", err)
	}
	return nil
}
