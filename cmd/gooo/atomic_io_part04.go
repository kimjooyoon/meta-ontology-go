package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func defaultAtomicFileOps() atomicFileOps {
	return atomicFileOps{createTemp: os.CreateTemp, syncFile: func(file *os.File) error { return file.Sync() }, rename: os.Rename, remove: os.Remove, syncDir: syncDirectory}
}
func writeAtomicFile(path string, data []byte) error {
	return writeAtomicFileWithOps(path, data, defaultAtomicFileOps())
}
func writeAtomicFileWithOps(path string, data []byte, ops atomicFileOps) error {
	if int64(len(data)) > maxOutputBytes {
		return outputLimitError(maxOutputBytes)
	}
	dir := filepath.Dir(path)
	temp, err := ops.createTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tempName := temp.Name()
	keepTemp := true
	defer func() {
		_ = temp.Close()
		if keepTemp {
			_ = ops.remove(tempName)
		}
	}()
	if err := prepareAtomicTemp(temp, data, ops); err != nil {
		return err
	}
	if err := validateOutputTarget(path); err != nil {
		return err
	}
	if err := ops.rename(tempName, path); err != nil {
		return fmt.Errorf("rename temporary output: %w", err)
	}
	keepTemp = false
	if err := ops.syncDir(dir); err != nil {
		return fmt.Errorf("sync output directory: %w", err)
	}
	return nil
}
func writeAll(file *os.File, data []byte) error {
	for len(data) > 0 {
		written, err := file.Write(data)
		if written > 0 {
			data = data[written:]
		}
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
