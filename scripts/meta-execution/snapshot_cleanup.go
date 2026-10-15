package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

// Only the parent returned by readOnlyGitSnapshot belongs to this cleanup.
// Directory write permission is needed to unlink children, even when the files
// themselves can remain read-only. Walk never follows symbolic links.
func removeReadOnlySnapshot(root string) error {
	prepareErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, info.Mode().Perm()|0o700)
		}
		return nil
	})
	return errors.Join(prepareErr, os.RemoveAll(root))
}

func cleanupGitSnapshot(root string) {
	if err := removeReadOnlySnapshot(root); err != nil {
		log.Printf("remove disposable Git snapshot %q: %v", root, err)
	}
}
