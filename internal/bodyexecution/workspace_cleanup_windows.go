package bodyexecution

import (
	"errors"
	"syscall"
)

// Win32 error identities; Unix errno numbers have different meanings.
// https://learn.microsoft.com/windows/win32/debug/system-error-codes--0-499-
const (
	windowsSharingViolation syscall.Errno = 32
	windowsLockViolation    syscall.Errno = 33
)

func retryableWorkspaceRemoval(err error) bool {
	return errors.Is(err, windowsSharingViolation) || errors.Is(err, windowsLockViolation)
}
