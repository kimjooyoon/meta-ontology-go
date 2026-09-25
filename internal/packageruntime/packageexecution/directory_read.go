package packageexecution

import (
	"io"
	"os"
)

func readSourceFile(filename string) ([]byte, error) {
	source, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	return readBoundedSource(source)
}

func readBoundedSource(source io.Reader) ([]byte, error) {
	// One sentinel byte lets LoadDirectory reject oversize inputs without
	// reading the rest of the file or trusting a potentially stale file size.
	return io.ReadAll(io.LimitReader(source, maxSourceBytes+1))
}
