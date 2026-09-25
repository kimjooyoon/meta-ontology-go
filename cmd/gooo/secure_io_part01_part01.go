package main

import (
	"fmt"
	"io"
	"os"
)

const (
	maxInputBytes     int64 = 1 << 20
	maxOutputBytes    int64 = 16 << 20
	generatedFileName       = "semantic.gooo.go"
)

func readSource(reader SourceReader, filename string) ([]byte, error) {
	source, err := reader.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if int64(len(source)) > maxInputBytes {
		return nil, inputLimitError(maxInputBytes)
	}
	return source, nil
}
func readRegularFile(path string, limit int64) ([]byte, error) {
	file, err := openRegularFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if info.Size() > limit {
		return nil, inputLimitError(limit)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	if int64(len(data)) > limit {
		return nil, inputLimitError(limit)
	}
	return data, nil
}
func inputLimitError(limit int64) error {
	return fmt.Errorf("input exceeds maximum size of %d bytes", limit)
}
func validateRegularFile(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%q is a symbolic link", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%q is not a regular file", path)
	}
	return nil
}
