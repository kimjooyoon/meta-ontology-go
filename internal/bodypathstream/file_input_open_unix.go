//go:build unix

package bodypathstream

import (
	"os"
	"syscall"
)

// A replaced FIFO must not wait before the opened descriptor is inspected.
func openFileInput(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
