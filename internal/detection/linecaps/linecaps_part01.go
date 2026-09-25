package linecaps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const (
	// DefaultMaxFileLines is the DAMP file cap used by this repository.
	DefaultMaxFileLines = sourcepolicy.DefaultMaxFileLines
	// DefaultMaxFunctionLines is the DRY function cap used by this repository.
	DefaultMaxFunctionLines = sourcepolicy.DefaultMaxFunctionLines
)

// Limits contains the inclusive maximum sizes accepted by Analyze.
type Limits struct {
	MaxFileLines     int
	MaxFunctionLines int
}

// DefaultLimits returns the canonical repository line policy.
func DefaultLimits() Limits {
	return Limits{MaxFileLines: DefaultMaxFileLines, MaxFunctionLines: DefaultMaxFunctionLines}
}
func (l Limits) validate() error {
	if l.MaxFileLines <= 0 || l.MaxFunctionLines <= 0 {
		return fmt.Errorf("linecaps limits must be positive")
	}
	return nil
}

// Discover returns repository-relative Go paths in lexical order. .git and
// vendor directories are excluded so generated or vendored copies cannot
// silently expand the verification scope.
func Discover(root string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("linecaps root must not be empty")
	}
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
