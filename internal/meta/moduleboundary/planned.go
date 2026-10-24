package moduleboundary

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckPackageWriteSet binds a topology recipe to its existing subject module.
// Missing destination directories inherit their nearest existing ancestor's
// owner. The recipe may not change the module/workspace graph it was bound to.
// It checks the whole set before the caller creates directories or moves files.
func CheckPackageWriteSet(root, subject string, paths ...string) (string, error) {
	owner, err := Directory(root, subject)
	if err != nil {
		return "", err
	}
	if err := checkDirectoryLinks(root, subject); err != nil {
		return "", err
	}
	if supported, err := LocalToolchainSupported(owner); err != nil || !supported {
		return "", fmt.Errorf("local Go toolchain cannot partition module for %q", subject)
	}
	for _, path := range paths {
		if err := preserveModuleGraph(path); err != nil {
			return "", err
		}
		targetOwner, err := plannedSourceOwner(root, path)
		if err != nil {
			return "", err
		}
		if targetOwner != owner {
			return "", fmt.Errorf("package write set crosses Go module boundary: %s", path)
		}
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(canonicalRoot, owner)
	return filepath.ToSlash(relative), err
}

func plannedSourceOwner(root, path string) (string, error) {
	if _, err := os.Lstat(path); err == nil {
		return Source(root, path)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	for directory := filepath.Dir(path); ; directory = filepath.Dir(directory) {
		if _, err := os.Lstat(directory); err == nil {
			owner, err := Directory(root, directory)
			if err != nil {
				return "", err
			}
			return owner, checkDirectoryLinks(root, directory)
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if directory == filepath.Dir(directory) {
			return "", fmt.Errorf("planned source has no existing parent: %s", path)
		}
	}
}

func preserveModuleGraph(path string) error {
	if isModuleGraphFile(path) {
		return fmt.Errorf("GO_MODULE_MIGRATION_CONTRACT_REQUIRED: %s", path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		if isModuleGraphFile(resolved) {
			return fmt.Errorf("GO_MODULE_MIGRATION_CONTRACT_REQUIRED: %s", path)
		}
		return nil
	}
	if _, statErr := os.Lstat(path); statErr == nil || !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isModuleGraphFile(path string) bool {
	switch filepath.Base(path) {
	case "go.mod", "go.sum", "go.work", "go.work.sum":
		return true
	default:
		return false
	}
}
