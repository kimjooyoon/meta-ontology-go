package verify

import (
	"fmt"
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

// LinePolicy defines deterministic source constraints used by CI.
type LinePolicy struct {
	MaxFileLines         int
	MaxFunctionLines     int
	MaxDirectDirectoryIn int
}

// DefaultLinePolicy returns the currently active repository constraints.
func DefaultLinePolicy() LinePolicy {
	return LinePolicy{
		MaxFileLines:         75,
		MaxFunctionLines:     75,
		MaxDirectDirectoryIn: 10,
	}
}

// CheckGoCaps checks the DAMP file limit and DRY function limit. If files is
// empty, all source files below root are discovered in lexical path order.
func CheckGoCaps(root string, files []string, maxFileLines, maxFunctionLines int) error {
	return CheckSourcePolicy(root, files, LinePolicy{
		MaxFileLines:     maxFileLines,
		MaxFunctionLines: maxFunctionLines,
	})
}
