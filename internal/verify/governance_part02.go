package verify

import (
	"fmt"
	"strings"
)

// ValidateGovernanceMatrix keeps JSON policy and executable ownership aligned.
func ValidateGovernanceMatrix(matrix GovernanceMatrix) error {
	if matrix.Schema != GovernanceSchemaVersion {
		return fmt.Errorf("unsupported governance schema %q", matrix.Schema)
	}
	if matrix.Mode != "ci_only" {
		return fmt.Errorf("governance mode must be ci_only")
	}
	if !sameStrings(matrix.ProofJobs, canonicalJobs()) {
		return fmt.Errorf("proof jobs do not match canonical CI jobs")
	}
	if !sameStrings(matrix.RequiredContexts.Dev, canonicalJobs()) || !sameStrings(matrix.RequiredContexts.Main, canonicalJobs()) {
		return fmt.Errorf("required contexts must be the six canonical CI jobs")
	}
	if len(matrix.RequiredContexts.Dev) != len(canonicalJobs()) || len(matrix.RequiredContexts.Main) != len(canonicalJobs()) {
		return fmt.Errorf("required context names are not unique and canonical")
	}
	if err := validateGovernanceOwnership(matrix.Ownership); err != nil {
		return err
	}
	if err := validateProtectedPushBranches(matrix.ProtectedPushBranches); err != nil {
		return err
	}
	if err := validateKernelPaths(matrix.ProtectedKernel); err != nil {
		return err
	}
	return validatePromotion(matrix.Promotion)
}
func validateProtectedPushBranches(branches []string) error {
	if !sameStrings(branches, []string{"dev", "main"}) {
		return fmt.Errorf("protected push branches must be dev and main")
	}
	for _, branch := range branches {
		if branch == "" || strings.ContainsAny(branch, "/*?[]") {
			return fmt.Errorf("invalid protected push branch %q", branch)
		}
	}
	return nil
}
func validateGovernanceOwnership(ownership []GovernanceOwnership) error {
	if len(ownership) != len(branchScopeAllowlist) {
		return fmt.Errorf("governance ownership must contain every executable branch")
	}
	seen := make(map[string]bool, len(ownership))
	for _, entry := range ownership {
		if seen[entry.Branch] || strings.ContainsAny(entry.Branch, "*?[]") {
			return fmt.Errorf("duplicate or wildcard ownership branch %q", entry.Branch)
		}
		seen[entry.Branch] = true
		expected, ok := branchScopeAllowlist[entry.Branch]
		if !ok || !samePaths(expected, entry.Paths) {
			return fmt.Errorf("governance ownership mismatch for %q", entry.Branch)
		}
	}
	for branch := range branchScopeAllowlist {
		if !seen[branch] {
			return fmt.Errorf("governance ownership missing branch %q", branch)
		}
	}
	return nil
}
