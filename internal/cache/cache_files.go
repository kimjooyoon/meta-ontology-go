package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func cleanupShard(shardPath string) error {
	info, err := os.Lstat(shardPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil
	}
	entries, err := os.ReadDir(shardPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !isTemporaryName(entry.Name()) {
			continue
		}
		path := filepath.Join(shardPath, entry.Name())
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if err := removeCacheEntry(path, info); err != nil {
			return fmt.Errorf("remove %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func isShardName(name string) bool {
	return len(name) == 2 && isLowerHex(name)
}

func isDigestName(name string) bool {
	return len(name) == digestLength && isLowerHex(name)
}

func isLowerHex(name string) bool {
	for _, char := range name {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func isTemporaryName(name string) bool {
	return strings.HasPrefix(name, ".") && strings.Contains(name, ".tmp-")
}
