package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
	if err != nil || !metadataSane(metadata) || !metadataBindsToEntry(metadata, shardPath, entry.Name()) ||
		!filter.matches(metadata) {
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
