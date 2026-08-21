package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
	if err != nil || !metadataSane(metadata) || !metadataBindsToEntry(metadata, shardPath, entry.Name()) ||
		!filter.matches(metadata) {
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
func metadataBindsToEntry(metadata Metadata, shardPath, entryName string) bool {
	return metadata.Key == entryName && len(entryName) >= 2 && filepath.Base(shardPath) == entryName[:2]
}
