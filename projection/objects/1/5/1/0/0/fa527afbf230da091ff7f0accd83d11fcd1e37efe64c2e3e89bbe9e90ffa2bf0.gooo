package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Clear removes every cache object, including stale temporary directories.
func (c *Cache) Clear() error {
	c.filesystemMu.Lock()
	defer c.filesystemMu.Unlock()
	entries, err := os.ReadDir(c.objects)
	if errors.Is(err, os.ErrNotExist) {
		return ensureDirectory(c.objects, 0o700)
	}
	if err != nil {
		return fmt.Errorf("cache: list objects: %w", err)
	}
	for _, entry := range entries {
		path := filepath.Join(c.objects, entry.Name())
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if err := removeCacheEntry(path, info); err != nil {
			return fmt.Errorf("cache: clear %s: %w", entry.Name(), err)
		}
	}
	return nil
}
func (filter InvalidationFilter) validate() error {
	if filter.Namespace == "" && filter.KeyVersion == "" && filter.ToolVersion == "" &&
		filter.ArtifactType == "" && filter.Projection == "" {
		return ErrEmptyFilter
	}
	fields := []struct {
		label string
		value string
	}{
		{"namespace", filter.Namespace}, {"key version", filter.KeyVersion},
		{"tool version", filter.ToolVersion}, {"artifact type", filter.ArtifactType},
		{"projection", filter.Projection},
	}
	for _, field := range fields {
		if err := validateKeyComponent(field.label, field.value, false); err != nil {
			return err
		}
	}
	return nil
}
func (filter InvalidationFilter) matches(metadata Metadata) bool {
	return (filter.Namespace == "" || filter.Namespace == metadata.Namespace) &&
		(filter.KeyVersion == "" || filter.KeyVersion == metadata.KeyVersion) &&
		(filter.ToolVersion == "" || filter.ToolVersion == metadata.ToolVersion) &&
		(filter.ArtifactType == "" || filter.ArtifactType == metadata.ArtifactType) &&
		(filter.Projection == "" || filter.Projection == metadata.Projection)
}
