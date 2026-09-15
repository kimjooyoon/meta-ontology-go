package verify

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// Violation is one deterministic policy failure.
type Violation struct {
	Path   string
	Rule   string
	Actual int
	Limit  int
	Detail string
}

func (v Violation) Error() string {
	if v.Detail != "" {
		return fmt.Sprintf("%s: %s: %s", v.Path, v.Rule, v.Detail)
	}
	return fmt.Sprintf("%s: %s: got %d, limit %d", v.Path, v.Rule, v.Actual, v.Limit)
}

// LinePolicy aliases the canonical meta-policy used by metrics and tools.
type LinePolicy = sourcepolicy.Policy

// DefaultLinePolicy returns the currently active repository constraints.
func DefaultLinePolicy() LinePolicy {
	return sourcepolicy.Default()
}

// CheckGoCaps checks the DAMP file limit and DRY function limit. If files is
// empty, all source files below root are discovered in lexical path order.
func CheckGoCaps(root string, files []string, maxFileLines, maxFunctionLines int) error {
	return CheckSourcePolicy(root, files, LinePolicy{
		MaxFileLines:     maxFileLines,
		MaxFunctionLines: maxFunctionLines,
	})
}
