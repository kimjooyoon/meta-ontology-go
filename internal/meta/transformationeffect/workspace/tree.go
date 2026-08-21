package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func Scan(root string) (State, error) {
	state := State{Entries: []Entry{}}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "." || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, kind := []byte{}, "file"
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			data, kind = []byte(target), "symlink"
		} else if info.Mode().IsRegular() {
			data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported workspace entry %s", rel)
		}
		state.Entries = append(state.Entries, Entry{rel, kind, uint32(info.Mode()), hashBytes(data), data})
		return nil
	})
	if err != nil {
		return state, err
	}
	sort.Slice(state.Entries, func(i, j int) bool { return state.Entries[i].Path < state.Entries[j].Path })
	state.Digest = hashJSON(state.Entries)
	return state, nil
}
