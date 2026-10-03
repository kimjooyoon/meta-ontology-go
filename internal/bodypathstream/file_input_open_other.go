//go:build !unix

package bodypathstream

import (
	"fmt"
	"os"
)

func openFileInput(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: input must be a regular file", path)
	}
	return os.Open(path)
}
