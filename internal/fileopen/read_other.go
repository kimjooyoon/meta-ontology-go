//go:build !unix

package fileopen

import (
	"fmt"
	"os"
)

// ReadOnly retains the preliminary regular-file check on other platforms.
// Callers must still inspect the opened descriptor and enforce their bounds.
func ReadOnly(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: input must be a regular file", path)
	}
	return os.Open(path)
}
