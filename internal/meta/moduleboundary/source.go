package moduleboundary

import (
	"fmt"
	"os"
	"path/filepath"
)

// Source retains the logical owner when resolving a source link. A new file
// may not exist yet, but its parent directory must already have an owner.
func Source(root, path string) (string, error) {
	owner, err := Directory(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	if err := checkDirectoryLinks(root, filepath.Dir(path)); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if os.IsNotExist(err) {
		return owner, nil
	}
	if err != nil {
		return "", err
	}
	target, err := Directory(root, filepath.Dir(resolved))
	if err != nil || target != owner {
		return "", fmt.Errorf("source crosses Go module boundary: %s", path)
	}
	return owner, nil
}

func checkDirectoryLinks(root, directory string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	current, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	for {
		if current == root || current == resolvedRoot {
			return nil
		}
		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			return err
		}
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		parent := filepath.Dir(current)
		if info.Mode()&os.ModeSymlink != 0 {
			from, fromErr := Directory(root, parent)
			to, toErr := Directory(root, resolved)
			if fromErr != nil || toErr != nil || from != to {
				return fmt.Errorf("directory link crosses Go module boundary: %s", current)
			}
		}
		if resolved == resolvedRoot {
			return nil
		}
		if parent == current {
			return fmt.Errorf("source directory escapes repository: %s", directory)
		}
		current = parent
	}
}

// CheckWriteSet requires one owning module and a compatible local toolchain.
// It performs no writes and does not borrow another module's dependencies.
func CheckWriteSet(root, source string, targets ...string) error {
	owner, err := Source(root, source)
	if err != nil {
		return err
	}
	if supported, err := LocalToolchainSupported(owner); err != nil || !supported {
		return fmt.Errorf("local Go toolchain cannot rewrite module for %q", source)
	}
	for _, target := range targets {
		targetOwner, err := Source(root, target)
		if err != nil {
			return err
		}
		if targetOwner != owner {
			return fmt.Errorf("write set crosses Go module boundary: %s -> %s", source, target)
		}
	}
	return nil
}
