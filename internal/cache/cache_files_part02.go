package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
