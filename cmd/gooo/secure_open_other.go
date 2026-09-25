//go:build !darwin && !linux

package main

import (
	"fmt"
	"os"
)

func openRegularFile(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err := validateRegularFile(path, info); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	openedInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if err := validateRegularFile(path, openedInfo); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}
