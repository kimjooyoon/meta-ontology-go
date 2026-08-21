package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func validateDirectoryProjection(source string, generated []generatedFile, maxEntries int) error {
	entries, err := os.ReadDir(filepath.Dir(source))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("separate-directory-kinds required before splitting %s", source)
		}
	}
	projected := len(entries) - 1 + len(generated)
	if projected > maxEntries {
		return fmt.Errorf("partition-directory required for %s: projected entries %d exceed %d", source, projected, maxEntries)
	}
	return nil
}

func commitGenerated(source string, generated []generatedFile) error {
	staged := make([]generatedFile, 0, len(generated))
	for _, output := range generated {
		if _, err := os.Lstat(output.path); err == nil {
			cleanupStaged(staged)
			return fmt.Errorf("refusing to overwrite generated path: %s", output.path)
		} else if !os.IsNotExist(err) {
			cleanupStaged(staged)
			return err
		}
		temp, err := os.CreateTemp(filepath.Dir(output.path), ".source-split-*")
		if err != nil {
			cleanupStaged(staged)
			return err
		}
		name := temp.Name()
		if _, err = temp.Write(output.contents); err == nil {
			err = temp.Chmod(0o644)
		}
		if closeErr := temp.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(name)
			cleanupStaged(staged)
			return err
		}
		staged = append(staged, generatedFile{path: name, contents: []byte(output.path)})
	}
	return promoteStaged(source, staged)
}
