package verify

import (
	"fmt"
	"sort"
	"strings"
)

// CheckSourcePolicy checks all repository-wide source constraints from policy.
func CheckSourcePolicy(root string, files []string, policy LinePolicy) error {
	if policy.MaxFileLines <= 0 || policy.MaxFunctionLines <= 0 {
		return fmt.Errorf("source policy caps must be positive")
	}
	if len(files) == 0 {
		var err error
		files, err = discoverSourceFiles(root)
		if err != nil {
			return err
		}
	}
	violations := make([]Violation, 0)
	for _, path := range sortedUnique(files) {
		violations = append(violations, checkSourceFile(root, path, policy)...)
	}
	if policy.MaxDirectDirectoryIn > 0 || policy.RequireHomogeneousDirectories {
		violations = append(violations, checkDirectoryLayout(root, policy)...)
	}
	if len(violations) == 0 {
		return nil
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Path != violations[j].Path {
			return violations[i].Path < violations[j].Path
		}
		return violations[i].Rule < violations[j].Rule
	})
	lines := make([]string, len(violations))
	for i, violation := range violations {
		lines[i] = violation.Error()
	}
	return fmt.Errorf("source policy failed:\n%s", strings.Join(lines, "\n"))
}
