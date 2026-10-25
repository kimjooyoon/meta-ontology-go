package main

import (
	"errors"
	"fmt"
	"os"
)

type atomicFileOps struct {
	createTemp    func(string, string) (*os.File, error)
	syncFile      func(*os.File) error
	rename        func(string, string) error
	remove        func(string) error
	syncDir       func(string) error
	beforePublish func() error
}
type atomicWrite struct {
	path   string
	data   []byte
	before *outputSnapshot
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
	if err := checkAtomicPreconditions(writes); err != nil {
		return err
	}
	if ops.beforePublish != nil {
		if err := ops.beforePublish(); err != nil {
			return err
		}
	}
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
		if write.before != nil && !sameOutputSnapshot(snapshot, *write.before) {
			return nil, nil, fmt.Errorf("atomic input changed: %q", write.path)
		}
		changed[index] = !snapshot.exists || !bytesEqual(snapshot.data, write.data)
	}
	return snapshots, changed, nil
}

func sameOutputSnapshot(left, right outputSnapshot) bool {
	return left.path == right.path && left.exists == right.exists &&
		(!left.exists || left.mode.Perm() == right.mode.Perm() && bytesEqual(left.data, right.data))
}

func checkAtomicPreconditions(writes []atomicWrite) error {
	for _, write := range writes {
		if write.before == nil {
			continue
		}
		current, err := captureOutputSnapshot(write.path)
		if err != nil {
			return err
		}
		if !sameOutputSnapshot(current, *write.before) {
			return fmt.Errorf("atomic input changed: %q", write.path)
		}
	}
	return nil
}
