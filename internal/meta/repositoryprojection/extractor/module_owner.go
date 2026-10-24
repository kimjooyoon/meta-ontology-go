package extractor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/moduleboundary"
)

// sourceModuleRoot chooses the nearest go.mod without changing repository-relative
// source and generated-file identities.
func sourceModuleRoot(root, logical string) (string, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return "", err
	}
	owner, err := moduleboundary.Directory(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if os.IsNotExist(err) {
		return owner, nil // Parsed sources may be virtual during type evidence construction.
	}
	if err != nil {
		return "", err
	}
	targetOwner, err := moduleboundary.Directory(root, filepath.Dir(resolved))
	if err != nil || targetOwner != owner {
		return "", fmt.Errorf("source crosses module boundary: %s", logical)
	}
	return owner, nil
}
