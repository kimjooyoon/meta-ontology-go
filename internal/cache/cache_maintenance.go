package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InvalidateKey removes one exact cache object. It is safe when absent.
func (c *Cache) InvalidateKey(key Key) (bool, error) {
	if err := validateFullKey(key); err != nil {
		return false, err
	}
	release, err := c.acquireKey(key)
	if err != nil {
		return false, err
	}
	defer release()
	c.filesystemMu.Lock()
	defer c.filesystemMu.Unlock()
	path, _ := c.objectPath(key)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache: inspect %s: %w", key, err)
	}
	if err := removeCacheEntry(path, info); err != nil {
		return false, fmt.Errorf("cache: invalidate %s: %w", key, err)
	}
	return true, nil
}

// Invalidate removes entries matching filter and returns the number removed.
// Corrupt metadata is skipped because it cannot be matched reliably.
func (c *Cache) Invalidate(filter InvalidationFilter) (int, error) {
	if err := filter.validate(); err != nil {
		return 0, err
	}
	c.filesystemMu.Lock()
	defer c.filesystemMu.Unlock()
	shards, err := os.ReadDir(c.objects)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("cache: list objects: %w", err)
	}
	removed := 0
	for _, shard := range shards {
		count, err := c.invalidateShard(shard, filter)
		if err != nil {
			return removed, err
		}
		removed += count
	}
	return removed, nil
}

func (c *Cache) invalidateShard(shard os.DirEntry, filter InvalidationFilter) (int, error) {
	if !isShardName(shard.Name()) {
		return 0, nil
	}
	shardPath := filepath.Join(c.objects, shard.Name())
	shardInfo, err := os.Lstat(shardPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	if shardInfo.Mode()&os.ModeSymlink != 0 || !shardInfo.IsDir() {
		return 0, nil
	}
	entries, err := os.ReadDir(shardPath)
	if err != nil {
		return 0, fmt.Errorf("cache: list shard %s: %w", shard.Name(), err)
	}
	removed := 0
	for _, entry := range entries {
		if !isDigestName(entry.Name()) {
			continue
		}
		matched, err := c.invalidateEntry(shardPath, entry, filter)
		if err != nil {
			return removed, err
		}
		if matched {
			removed++
		}
	}
	return removed, nil
}

func (c *Cache) invalidateEntry(shardPath string, entry os.DirEntry, filter InvalidationFilter) (bool, error) {
	entryPath := filepath.Join(shardPath, entry.Name())
	entryInfo, err := os.Lstat(entryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if entryInfo.Mode()&os.ModeSymlink != 0 || !entryInfo.IsDir() {
		return false, nil
	}
	metadata, err := readMetadataAt(filepath.Join(entryPath, metaFileName))
	if err != nil || !metadataSane(metadata) || !filter.matches(metadata) {
		return false, nil
	}
	if err := removeCacheEntry(entryPath, entryInfo); err != nil {
		return false, fmt.Errorf("cache: invalidate %s: %w", entry.Name(), err)
	}
	return true, nil
}

// InvalidateNamespace removes all projections for a semantic namespace.
func (c *Cache) InvalidateNamespace(namespace string) (int, error) {
	if err := validateKeyComponent("namespace", namespace, true); err != nil {
		return 0, err
	}
	return c.Invalidate(InvalidationFilter{Namespace: namespace})
}

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
