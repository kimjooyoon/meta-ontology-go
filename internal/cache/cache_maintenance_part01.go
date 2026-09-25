package cache

import (
	"errors"
	"fmt"
	"os"
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
