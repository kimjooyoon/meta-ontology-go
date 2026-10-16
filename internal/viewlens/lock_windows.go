package viewlens

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

var lockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")

func tryLock(file *os.File) (bool, error) {
	var overlap syscall.Overlapped
	result, _, err := lockFileEx.Call(file.Fd(), 3, 0, 1, 0, uintptr(unsafe.Pointer(&overlap)))
	runtime.KeepAlive(&overlap)
	if result != 0 {
		return true, nil
	}
	if err == syscall.Errno(33) {
		return false, nil
	}
	return false, fmt.Errorf("PROFILE_LOCK_FAILED: %w", err)
}
