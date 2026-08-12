//go:build windows

package cache

import (
	"os"
	"syscall"
)

func openReceiptRead(path string) (*os.File, error) {
	return openReceiptFile(path, syscall.GENERIC_READ, syscall.OPEN_EXISTING)
}

func openReceiptAppend(path string) (*os.File, error) {
	return openReceiptFile(path, syscall.GENERIC_WRITE, syscall.OPEN_ALWAYS)
}

func openReceiptFile(path string, access, disposition uint32) (*os.File, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(name, access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, disposition, syscall.FILE_ATTRIBUTE_NORMAL|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		_ = syscall.CloseHandle(handle)
		return nil, ErrUnsafeReceiptLog
	}
	if err := validateReceiptFile(file); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}
