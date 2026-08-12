//go:build !windows

package cache

import (
	"os"
	"syscall"
)

func openReceiptRead(path string) (*os.File, error) {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, ErrUnsafeReceiptLog
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	if err := validateReceiptFile(file); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func openReceiptAppend(path string) (*os.File, error) {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, ErrUnsafeReceiptLog
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	if err := validateReceiptFile(file); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}
