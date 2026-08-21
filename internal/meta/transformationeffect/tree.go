package transformationeffect

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fileState struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Mode   uint32 `json:"mode"`
	SHA256 string `json:"sha256"`
	data   []byte
}

type treeState struct {
	Entries []fileState
	Digest  string
}

func scanTree(root string) (treeState, error) {
	state := treeState{Entries: []fileState{}}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".git" || strings.HasPrefix(rel, ".git/") {
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
		state.Entries = append(state.Entries, fileState{rel, kind, uint32(info.Mode()), hashBytes(data), data})
		return nil
	})
	if err != nil {
		return state, err
	}
	sort.Slice(state.Entries, func(i, j int) bool { return state.Entries[i].Path < state.Entries[j].Path })
	state.Digest = hashJSON(state.Entries)
	return state, nil
}

func copyTree(state treeState, root string) error {
	for _, entry := range state.Entries {
		path := filepath.Join(root, filepath.FromSlash(entry.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if entry.Kind == "symlink" {
			if err := os.Symlink(string(entry.data), path); err != nil {
				return err
			}
		} else if err := os.WriteFile(path, entry.data, os.FileMode(entry.Mode).Perm()); err != nil {
			return err
		}
	}
	return nil
}
