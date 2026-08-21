package adapter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func validateFileObservation(observation FileObservation, requireFile bool) error {
	if strings.TrimSpace(observation.Path) == "" {
		return fmt.Errorf("path is required")
	}
	if !observation.Exists {
		if observation.Kind != "missing" || observation.ByteDigest != "" || observation.Lstat.Exists {
			return fmt.Errorf("missing path has contradictory state")
		}
		return nil
	}
	if !validKind(observation.Kind) || observation.Kind == "missing" {
		return fmt.Errorf("invalid path kind %q", observation.Kind)
	}
	if !observation.Lstat.Exists || observation.Lstat.Mode == "" {
		return fmt.Errorf("lstat identity is incomplete")
	}
	if requireFile && observation.Kind != "file" {
		return fmt.Errorf("primary path is not a regular file")
	}
	if observation.Kind == "file" || observation.Kind == "symlink" {
		if !validDigest(observation.ByteDigest) {
			return fmt.Errorf("byte digest is incomplete")
		}
	} else if observation.ByteDigest != "" {
		return fmt.Errorf("directory or special path has bytes")
	}
	return nil
}
func validateTempSnapshot(snapshot TempArtifactSnapshot) error {
	if snapshot.Root.Exists && snapshot.Root.Mode == "" {
		return fmt.Errorf("temp root lstat identity is incomplete")
	}
	if !snapshot.Root.Exists && snapshot.Root != (LstatIdentity{}) {
		return fmt.Errorf("missing temp root has contradictory identity")
	}
	if !validDigest(snapshot.Digest) {
		return fmt.Errorf("snapshot digest is incomplete")
	}
	entries := append([]FileObservation{}, snapshot.Entries...)
	if !sort.SliceIsSorted(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path }) {
		return fmt.Errorf("snapshot entries are not canonical")
	}
	for index, entry := range entries {
		if filepath.IsAbs(entry.Path) || entry.Path == "." || entry.Path == ".." || strings.HasPrefix(entry.Path, "../") {
			return fmt.Errorf("snapshot entry path escapes temp root")
		}
		if index > 0 && entries[index-1].Path == entry.Path {
			return fmt.Errorf("snapshot entry path is duplicated")
		}
		if err := validateFileObservation(entry, false); err != nil {
			return err
		}
	}
	computed, err := digestTempSnapshot(snapshot.Root, entries)
	if err != nil || computed != snapshot.Digest {
		return fmt.Errorf("snapshot digest does not match entries")
	}
	return nil
}
