package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (c *Cache) getLocked(key Key) ([]byte, Metadata, error) {
	if err := c.validatePathKey(key); err != nil {
		return nil, Metadata{}, err
	}
	c.filesystemMu.RLock()
	defer c.filesystemMu.RUnlock()
	return c.readObject(key)
}
func (c *Cache) putLocked(key Key, data []byte, info EntryInfo) error {
	if err := validateFullKey(key); err != nil {
		return err
	}
	if c.maxEntrySize > 0 && int64(len(data)) > c.maxEntrySize {
		return fmt.Errorf("cache: %w: %d > %d", ErrEntryTooLarge, len(data), c.maxEntrySize)
	}
	if err := validateEntryInfo(info); err != nil {
		return err
	}
	if info.Projection != "" && info.Projection != key.Projection {
		return fmt.Errorf("%w: EntryInfo projection is not key identity", ErrInvalidKey)
	}
	path, err := c.objectPath(key)
	if err != nil {
		return err
	}
	c.filesystemMu.Lock()
	defer c.filesystemMu.Unlock()
	return c.writeObject(path, key, data, info)
}
func (c *Cache) writeObject(path string, key Key, data []byte, info EntryInfo) error {
	if _, _, err := c.readObject(key); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrCorrupt) {
		return err
	}
	shardPath := filepath.Dir(path)
	if err := ensureDirectory(shardPath, 0o700); err != nil {
		return fmt.Errorf("cache: prepare shard: %w", err)
	}
	temporary, err := os.MkdirTemp(shardPath, "."+key.String()+".tmp-")
	if err != nil {
		return fmt.Errorf("cache: create staging directory: %w", err)
	}
	defer func() { _ = removeCacheEntryBestEffort(temporary) }()
	metadata := makeMetadata(key, data, info)
	if err := writeObjectFiles(temporary, data, metadata); err != nil {
		return err
	}
	bestEffortSyncDirectory(temporary)
	if err := commitObject(temporary, path, key, c); err != nil {
		return err
	}
	bestEffortSyncDirectory(shardPath)
	return nil
}
