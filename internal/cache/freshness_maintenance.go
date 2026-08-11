package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InvalidateStale removes scoped entries whose dependency or provenance
// identity differs from Current. Invalid entries are left for repair or
// Clear, because they cannot be classified as stale safely.
func (c *Cache) InvalidateStale(filter StaleFilter) (int, error) {
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
		count, err := c.invalidateStaleShard(shard, filter)
		if err != nil {
			return removed, err
		}
		removed += count
	}
	return removed, nil
}

func (c *Cache) invalidateStaleShard(shard os.DirEntry, filter StaleFilter) (int, error) {
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
		stale, err := c.invalidateStaleEntry(shardPath, entry, filter)
		if err != nil {
			return removed, err
		}
		if stale {
			removed++
		}
	}
	return removed, nil
}

func (c *Cache) invalidateStaleEntry(shardPath string, entry os.DirEntry, filter StaleFilter) (bool, error) {
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
	if metadataFreshness(metadata).Equal(filter.Current) {
		return false, nil
	}
	if err := removeCacheEntry(entryPath, entryInfo); err != nil {
		return false, fmt.Errorf("cache: invalidate stale %s: %w", entry.Name(), err)
	}
	return true, nil
}
