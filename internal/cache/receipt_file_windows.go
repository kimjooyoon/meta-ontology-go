//go:build windows

package cache

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	lockFileExProc   = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")
	unlockFileExProc = syscall.NewLazyDLL("kernel32.dll").NewProc("UnlockFileEx")
)

const lockFileExExclusive = 0x00000002

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

func acquireReceiptFileLock(path string) (func(), error) {
	file, err := openReceiptFile(path+".lock", syscall.GENERIC_READ|syscall.GENERIC_WRITE, syscall.OPEN_ALWAYS)
	if err != nil {
		return nil, err
	}
	var overlapped syscall.Overlapped
	result, _, callErr := lockFileExProc.Call(file.Fd(), lockFileExExclusive,
		0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
	if result == 0 {
		_ = file.Close()
		return nil, callErr
	}
	return func() {
		_, _, _ = unlockFileExProc.Call(file.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
		_ = file.Close()
	}, nil
}
