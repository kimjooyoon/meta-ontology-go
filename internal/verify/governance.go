package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const GovernanceSchemaVersion = "gooo/ci-governance/v1"

var canonicalGovernanceRoles = []string{"Builder", "Guardian", "Approver", "Gate"}

// GovernanceMatrix is the machine-readable CI trust and promotion contract.
type GovernanceMatrix struct {
	Schema          string                `json:"schema"`
	Roles           []GovernanceRole      `json:"roles"`
	Ownership       []GovernanceOwnership `json:"ownership"`
	ProtectedKernel []string              `json:"protected_kernel_paths"`
	Promotion       GovernancePromotion   `json:"promotion"`
}

type GovernanceRole struct {
	Name              string   `json:"name"`
	Independent       bool     `json:"independent"`
	MayApprove        bool     `json:"may_approve"`
	ForbiddenOverlaps []string `json:"forbidden_overlaps"`
}

type GovernanceOwnership struct {
	Branch string   `json:"branch"`
	Paths  []string `json:"paths"`
}

type GovernancePromotion struct {
	Source                      string   `json:"source"`
	Target                      string   `json:"target"`
	RequiredChecks              []string `json:"required_checks"`
	BranchProtectionRequired    bool     `json:"branch_protection_required"`
	IndependentGuardianApprover bool     `json:"independent_guardian_approver"`
	ForbidSelfApproval          bool     `json:"forbid_self_approval"`
	ForbidRoleOverlap           bool     `json:"forbid_role_overlap"`
}

// ReadGovernanceMatrix loads and validates the protected JSON contract.
func ReadGovernanceMatrix(filename string) (GovernanceMatrix, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return GovernanceMatrix{}, err
	}
	var matrix GovernanceMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		return GovernanceMatrix{}, fmt.Errorf("parse governance matrix: %w", err)
	}
	if err := ValidateGovernanceMatrix(matrix); err != nil {
		return GovernanceMatrix{}, err
	}
	return matrix, nil
}

// ValidateGovernanceMatrix keeps JSON policy and executable ownership aligned.
func ValidateGovernanceMatrix(matrix GovernanceMatrix) error {
	if matrix.Schema != GovernanceSchemaVersion {
		return fmt.Errorf("unsupported governance schema %q", matrix.Schema)
	}
	if err := validateGovernanceRoles(matrix.Roles); err != nil {
		return err
	}
	if err := validateGovernanceOwnership(matrix.Ownership); err != nil {
		return err
	}
	if err := validateKernelPaths(matrix.ProtectedKernel); err != nil {
		return err
	}
	return validatePromotion(matrix.Promotion)
}

func validateGovernanceRoles(roles []GovernanceRole) error {
	if len(roles) != len(canonicalGovernanceRoles) {
		return fmt.Errorf("governance matrix must define four roles")
	}
	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		if seen[role.Name] || !contains(canonicalGovernanceRoles, role.Name) {
			return fmt.Errorf("duplicate or unknown governance role %q", role.Name)
		}
		seen[role.Name] = true
		if role.Name == "Guardian" || role.Name == "Approver" {
			if !role.Independent {
				return fmt.Errorf("%s must be independent", role.Name)
			}
		}
	}
	for _, name := range canonicalGovernanceRoles {
		if !seen[name] {
			return fmt.Errorf("missing governance role %q", name)
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

func validateKernelPaths(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("protected kernel paths are required")
	}
	for _, path := range paths {
		if path == "" || strings.ContainsAny(path, "*?[]") || strings.HasPrefix(path, "/") {
			return fmt.Errorf("invalid protected kernel path %q", path)
		}
	}
	return nil
}

func validatePromotion(promotion GovernancePromotion) error {
	if promotion.Source != "integration" || promotion.Target != "main" {
		return fmt.Errorf("promotion must be integration to main")
	}
	if !promotion.BranchProtectionRequired || !promotion.IndependentGuardianApprover || !promotion.ForbidSelfApproval || !promotion.ForbidRoleOverlap {
		return fmt.Errorf("promotion safeguards are incomplete")
	}
	if !sameStrings(promotion.RequiredChecks, canonicalJobs()) {
		return fmt.Errorf("promotion checks do not match canonical CI jobs")
	}
	return nil
}

func samePaths(left, right []string) bool {
	return sameStrings(normalizeMatrixPaths(left), normalizeMatrixPaths(right))
}

func normalizeMatrixPaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for index, path := range paths {
		normalized[index] = strings.TrimSuffix(path, "/**")
	}
	sort.Strings(normalized)
	return normalized
}

func sameStrings(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func canonicalJobs() []string {
	return []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}
}
