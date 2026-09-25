package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ensureDirectory(path string, permission os.FileMode) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, permission); err != nil {
			return err
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is a symbolic link", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}
func removeCacheEntry(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return os.Remove(path)
	}
	return os.RemoveAll(path)
}
func removeCacheEntryBestEffort(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return removeCacheEntry(path, info)
}
func bestEffortSyncDirectory(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	_ = file.Sync()
	_ = file.Close()
}
func cleanupTemporaryEntries(objects string) error {
	shards, err := os.ReadDir(objects)
	if err != nil {
		return err
	}
	for _, shard := range shards {
		if !isShardName(shard.Name()) {
			continue
		}
		if err := cleanupShard(filepath.Join(objects, shard.Name())); err != nil {
			return err
		}
	}
	return nil
}
