package workspace

import (
	"os"
	"path/filepath"
)

func copyTree(state State, root string) error {
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
