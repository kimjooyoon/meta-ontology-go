package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func commitObject(temporary, path string, key Key, cache *Cache) error {
	if err := os.Rename(temporary, path); err == nil {
		return nil
	} else if _, _, existingErr := cache.readObject(key); existingErr == nil {
		return nil
	}
	if info, statErr := os.Lstat(path); statErr == nil {
		if removeErr := removeCacheEntry(path, info); removeErr != nil {
			return fmt.Errorf("cache: replace corrupt entry: %w", removeErr)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("cache: inspect existing entry: %w", statErr)
	}
	if err := os.Rename(temporary, path); err != nil {
		if _, _, existingErr := cache.readObject(key); existingErr == nil {
			return nil
		}
		return fmt.Errorf("cache: commit projection: %w", err)
	}
	return nil
}
func (c *Cache) readObject(key Key) ([]byte, Metadata, error) {
	path, err := c.objectPath(key)
	if err != nil {
		return nil, Metadata{}, err
	}
	objectInfo, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, Metadata{}, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("cache: inspect %s: %w", key, err)
	}
	if objectInfo.Mode()&os.ModeSymlink != 0 || !objectInfo.IsDir() {
		return nil, Metadata{}, corruptError(key, "object is not a directory")
	}
	metadata, err := readMetadataAt(filepath.Join(path, metaFileName))
	if err != nil {
		return nil, Metadata{}, corruptError(key, "invalid metadata: "+err.Error())
	}
	if err := validateMetadataForKey(metadata, key); err != nil {
		return nil, Metadata{}, corruptError(key, err.Error())
	}
	data, err := readDataFile(filepath.Join(path, dataFileName), c.maxEntrySize)
	if err != nil {
		if errors.Is(err, ErrEntryTooLarge) {
			return nil, Metadata{}, err
		}
		return nil, Metadata{}, corruptError(key, "invalid projection: "+err.Error())
	}
	if int64(len(data)) != metadata.Size {
		return nil, Metadata{}, corruptError(key, fmt.Sprintf("size mismatch: metadata=%d actual=%d", metadata.Size, len(data)))
	}
	if HashBytes(data) != metadata.ContentDigest {
		return nil, Metadata{}, corruptError(key, "content digest mismatch")
	}
	return append([]byte(nil), data...), metadata, nil
}
