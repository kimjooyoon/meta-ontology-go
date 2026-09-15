package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// A new attempt must not inherit a previous terminal bundle if interrupted.
// Preserve the old bytes beside the caller-owned output, outside its active name.
// This is sequential storage hygiene, not a concurrent-writer lock or authority.
func archivePreviousObservation(bundlePath string) (string, error) {
	info, err := os.Lstat(bundlePath)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("observation output is not a regular file")
	}
	archive, err := os.CreateTemp(filepath.Dir(bundlePath), filepath.Base(bundlePath)+".previous-*")
	if err != nil {
		return "", err
	}
	name := archive.Name()
	if err := archive.Close(); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	if err := os.Rename(bundlePath, name); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	return name, nil
}
