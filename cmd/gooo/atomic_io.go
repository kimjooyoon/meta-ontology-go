package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type atomicFileOps struct {
	createTemp func(string, string) (*os.File, error)
	syncFile   func(*os.File) error
	rename     func(string, string) error
	remove     func(string) error
	syncDir    func(string) error
}

type atomicWrite struct {
	path string
	data []byte
}

type outputSnapshot struct {
	path   string
	exists bool
	mode   os.FileMode
	data   []byte
}

// writeAtomicFiles stages every artifact before publishing any of them. If a
// later publication or directory sync fails, already-published paths are
// restored from their exact pre-state snapshots.
func writeAtomicFiles(writes []atomicWrite) error {
	return writeAtomicFilesWithOps(writes, defaultAtomicFileOps())
}

func writeAtomicFilesWithOps(writes []atomicWrite, ops atomicFileOps) error {
	snapshots, changed, err := snapshotAtomicWrites(writes)
	if err != nil || !anyAtomicWriteChanged(changed) {
		return err
	}
	temps, err := stageAtomicWrites(writes, changed, ops)
	if err != nil {
		return err
	}
	defer removeAtomicTemps(temps, ops)
	return publishAtomicWrites(writes, changed, temps, snapshots, ops)
}

func snapshotAtomicWrites(writes []atomicWrite) ([]outputSnapshot, []bool, error) {
	snapshots := make([]outputSnapshot, len(writes))
	changed := make([]bool, len(writes))
	seen := make(map[string]struct{}, len(writes))
	for index, write := range writes {
		if write.path == "" {
			return nil, nil, errors.New("atomic output path is empty")
		}
		if _, exists := seen[write.path]; exists {
			return nil, nil, fmt.Errorf("duplicate atomic output path %q", write.path)
		}
		seen[write.path] = struct{}{}
		if int64(len(write.data)) > maxOutputBytes {
			return nil, nil, outputLimitError(maxOutputBytes)
		}
		if err := validateOutputTarget(write.path); err != nil {
			return nil, nil, fmt.Errorf("inspect output %q: %w", write.path, err)
		}
		snapshot, err := captureOutputSnapshot(write.path)
		if err != nil {
			return nil, nil, err
		}
		snapshots[index] = snapshot
		changed[index] = !snapshot.exists || !bytesEqual(snapshot.data, write.data)
	}
	return snapshots, changed, nil
}

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
		if err := prepareAtomicTemp(temp, write.data, ops); err != nil {
			_ = temp.Close()
			return nil, fmt.Errorf("prepare temporary output for %q: %w", write.path, err)
		}
	}
	cleanup = false
	return temps, nil
}

func prepareAtomicTemp(temp *os.File, data []byte, ops atomicFileOps) error {
	if err := temp.Chmod(0o644); err != nil {
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

func publishAtomicWrites(writes []atomicWrite, changed []bool, temps []string, snapshots []outputSnapshot, ops atomicFileOps) error {
	committed := make([]int, 0, len(writes))
	for index, write := range writes {
		if !changed[index] {
			continue
		}
		if err := ops.rename(temps[index], write.path); err != nil {
			return atomicPublishError(writes, snapshots, committed, write.path, err)
		}
		temps[index] = ""
		committed = append(committed, index)
	}
	for directory := range atomicWriteDirectories(writes) {
		if err := ops.syncDir(directory); err != nil {
			return atomicPublishError(writes, snapshots, committed, directory, err)
		}
	}
	return nil
}

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
	for index := len(committed) - 1; index >= 0; index-- {
		writeIndex := committed[index]
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

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return err
	}
	return directory.Close()
}
