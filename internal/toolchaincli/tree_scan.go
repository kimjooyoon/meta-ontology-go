package toolchaincli

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func scanTree(root string) ([]treeRecord, error) {
	records := make([]treeRecord, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		record := treeRecord{Path: filepath.ToSlash(relative), Mode: uint32(info.Mode()),
			Size: info.Size(), ModifiedNanos: info.ModTime().UnixNano()}
		if info.Mode().IsRegular() {
			record.ContentDigest, err = digestFile(path)
		} else if info.Mode()&os.ModeSymlink != 0 {
			var target string
			target, err = os.Readlink(path)
			record.ContentDigest = digestJSON(target)
		} else {
			record.ContentDigest = digestJSON(info.Mode().String())
		}
		if err == nil {
			records = append(records, record)
		}
		return err
	})
	sort.Slice(records, func(left, right int) bool { return records[left].Path < records[right].Path })
	return records, err
}
