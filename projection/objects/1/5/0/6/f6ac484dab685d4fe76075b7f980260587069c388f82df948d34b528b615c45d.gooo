//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"syscall"
)

func openRegularFile(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err := validateRegularFile(path, info); err != nil {
		return nil, err
	}
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, fmt.Errorf("open %q without following links: %w", path, err)
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("open %q: invalid file descriptor", path)
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
