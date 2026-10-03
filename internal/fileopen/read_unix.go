//go:build unix

package fileopen

import (
	"os"
	"syscall"
)

// ReadOnly opens without waiting for a FIFO writer. The caller must validate
// the opened descriptor's kind and byte bounds before reading any content.
func ReadOnly(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
