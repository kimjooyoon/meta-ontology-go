package writeset

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func SnapshotDirectory(root string) (Snapshot, error) {
	absolute, err := filepath.Abs(root)
	if err != nil { return Snapshot{}, err }
	entries := make([]Entry, 0)
	err = filepath.WalkDir(absolute, func(filePath string, item os.DirEntry, walkErr error) error {
		if walkErr != nil { return walkErr }
		relative, err := filepath.Rel(absolute, filePath)
		if err != nil { return err }
		if item.IsDir() {
			if relative != "." && strings.Split(filepath.ToSlash(relative), "/")[0] == ".git" { return filepath.SkipDir }
			return nil
		}
		entry, keep, err := snapshotEntry(filePath, filepath.ToSlash(relative), item)
		if err != nil { return err }
		if keep { entries = append(entries, entry) }
		return nil
	})
	if err != nil { return Snapshot{}, err }
	sort.Slice(entries, func(left, right int) bool { return entries[left].Path < entries[right].Path })
	return Snapshot{Schema: SnapshotSchema, RootDigest: digestEntries(entries), Entries: entries}, nil
}

func snapshotEntry(filePath, relative string, item os.DirEntry) (Entry, bool, error) {
	if item.Type()&os.ModeSymlink != 0 {
		target, err := os.Readlink(filePath)
		if err != nil { return Entry{}, false, err }
		return Entry{Path: relative, Kind: "SYMLINK", Size: int64(len(target)), Digest: digestBytes([]byte(target))}, true, nil
	}
	if !item.Type().IsRegular() { return Entry{}, false, nil }
	data, err := os.ReadFile(filePath)
	if err != nil { return Entry{}, false, err }
	return Entry{Path: relative, Kind: "FILE", Size: int64(len(data)), Digest: digestBytes(data)}, true, nil
}
