//go:build unix

package bodyexecution

import (
	"os"
	"syscall"
)

// Opening a replaced FIFO must not wait before the regular-file check runs.
func openDigestFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
