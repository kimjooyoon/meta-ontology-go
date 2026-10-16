//go:build darwin || linux

package viewlens

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func tryLock(file *os.File) (bool, error) {
	descriptor := file.Fd()
	if descriptor > uintptr(^uint(0)>>1) {
		return false, fmt.Errorf("PROFILE_LOCK_DESCRIPTOR_BOUND")
	}
	err := syscall.Flock(int(descriptor), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	return err == nil, err
}
