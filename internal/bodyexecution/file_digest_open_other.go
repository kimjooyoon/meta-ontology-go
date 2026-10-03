//go:build !unix

package bodyexecution

import "os"

func openDigestFile(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errNonRegularExecutable
	}
	return os.Open(path)
}
