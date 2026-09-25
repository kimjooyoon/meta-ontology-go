package verify

import (
	"sort"
)

// ConfiguredBranches returns branch names in deterministic order for policy
// diagnostics and tests.
func ConfiguredBranches() []string {
	branches := make([]string, 0, len(branchScopeAllowlist))
	for branch := range branchScopeAllowlist {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}
