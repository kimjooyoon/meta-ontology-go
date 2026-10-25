package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func anyAtomicWriteChanged(changed []bool) bool {
	for _, value := range changed {
		if value {
			return true
		}
	}
	return false
}
func stageAtomicWrites(writes []atomicWrite, changed []bool, ops atomicFileOps) ([]string, error) {
	temps := make([]string, len(writes))
	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		for index, name := range temps {
			if name != "" {
				_ = ops.remove(name)
				temps[index] = ""
			}
		}
	}()
	for index, write := range writes {
		if !changed[index] {
			continue
		}
		temp, err := ops.createTemp(filepath.Dir(write.path), "."+filepath.Base(write.path)+".tmp-*")
		if err != nil {
			return nil, fmt.Errorf("create temporary output for %q: %w", write.path, err)
		}
		temps[index] = temp.Name()
		mode := os.FileMode(0o644)
		if write.before != nil {
			mode = write.before.mode.Perm()
		}
		if err := prepareAtomicTempMode(temp, write.data, ops, mode); err != nil {
			_ = temp.Close()
			return nil, fmt.Errorf("prepare temporary output for %q: %w", write.path, err)
		}
	}
	cleanup = false
	return temps, nil
}
func prepareAtomicTemp(temp *os.File, data []byte, ops atomicFileOps) error {
	return prepareAtomicTempMode(temp, data, ops, 0o644)
}

func prepareAtomicTempMode(temp *os.File, data []byte, ops atomicFileOps, mode os.FileMode) error {
	if err := temp.Chmod(mode); err != nil {
		return fmt.Errorf("set mode: %w", err)
	}
	if err := writeAll(temp, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := ops.syncFile(temp); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return temp.Close()
}
