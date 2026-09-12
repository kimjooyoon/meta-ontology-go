package verify

import (
	"fmt"
	"sort"
	"strings"
)

// UnknownAgentBranchError retains denial while exposing scope-only hints.
// Candidates do not establish branch availability or any other authorization.
type UnknownAgentBranchError struct {
	branch     string
	candidates []string
}

func (failure *UnknownAgentBranchError) RequestedBranch() string {
	return failure.branch
}

func (failure *UnknownAgentBranchError) CandidateBranches() []string {
	return append([]string{}, failure.candidates...)
}

func (failure *UnknownAgentBranchError) Error() string {
	message := fmt.Sprintf("unknown agent branch %q; no paths are allowed", failure.branch)
	if len(failure.candidates) == 0 {
		return message
	}
	return fmt.Sprintf("%s; scope-only candidate branches: %s (branch availability and other authorization not evaluated)", message, strings.Join(failure.candidates, ", "))
}

func newUnknownAgentBranchError(branch string, paths []string) *UnknownAgentBranchError {
	return &UnknownAgentBranchError{branch: branch, candidates: CandidateScopeBranches(paths)}
}

// CandidateScopeBranches projects the existing ownership map, not new grants.
// Empty input and root placeholders cannot establish a changed-file scope.
func CandidateScopeBranches(paths []string) []string {
	candidates := []string{}
	if len(paths) == 0 {
		return candidates
	}
	paths = append([]string(nil), paths...)
	for _, path := range paths {
		if path == "." {
			return candidates
		}
	}
	for branch, allowed := range branchScopeAllowlist {
		if CheckPathScope(paths, allowed) == nil {
			candidates = append(candidates, branch)
		}
	}
	sort.Strings(candidates)
	return candidates
}
