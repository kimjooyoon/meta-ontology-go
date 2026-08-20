package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

func atomicPublishError(writes []atomicWrite, snapshots []outputSnapshot, committed []int, target string, publishErr error) error {
	if rollbackErr := rollbackAtomicWrites(writes, snapshots, committed); rollbackErr != nil {
		return fmt.Errorf("publish %q: %w; rollback failed: %v", target, publishErr, rollbackErr)
	}
	return fmt.Errorf("publish %q: %w", target, publishErr)
}
func atomicWriteDirectories(writes []atomicWrite) map[string]struct{} {
	directories := make(map[string]struct{}, len(writes))
	for _, write := range writes {
		directories[filepath.Dir(write.path)] = struct{}{}
	}
	return directories
}
func removeAtomicTemps(temps []string, ops atomicFileOps) {
	for _, name := range temps {
		if name != "" {
			_ = ops.remove(name)
		}
	}
}
func captureOutputSnapshot(path string) (outputSnapshot, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return outputSnapshot{path: path}, nil
	}
	if err != nil {
		return outputSnapshot{}, fmt.Errorf("inspect output %q: %w", path, err)
	}
	if err := validateRegularFile(path, info); err != nil {
		return outputSnapshot{}, err
	}
	data, err := readRegularFile(path, maxOutputBytes)
	if err != nil {
		return outputSnapshot{}, fmt.Errorf("snapshot output %q: %w", path, err)
	}
	return outputSnapshot{path: path, exists: true, mode: info.Mode(), data: data}, nil
}
func rollbackAtomicWrites(writes []atomicWrite, snapshots []outputSnapshot, committed []int) error {
	var firstErr error
	for _, writeIndex := range slices.Backward(committed) {

		snapshot := snapshots[writeIndex]
		var err error
		if snapshot.exists {
			err = writeAtomicFile(snapshot.path, snapshot.data)
			if err == nil {
				err = os.Chmod(snapshot.path, snapshot.mode.Perm())
			}
		} else {
			err = os.Remove(writes[writeIndex].path)
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
		}
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("restore %q: %w", writes[writeIndex].path, err)
		}
	}
	return firstErr
}
