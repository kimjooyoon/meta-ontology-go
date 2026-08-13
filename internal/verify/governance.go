package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const GovernanceSchemaVersion = "gooo/ci-governance/v2"

// GovernanceMatrix is the machine-readable CI trust and promotion contract.
type GovernanceMatrix struct {
	Schema                string                `json:"schema"`
	Mode                  string                `json:"mode"`
	CIAppID               int64                 `json:"ci_app_id"`
	RequiredContexts      GovernanceContexts    `json:"required_contexts"`
	GuardianContexts      GuardianContexts      `json:"guardian_contexts"`
	ProofJobs             []string              `json:"proof_jobs"`
	ProtectedPushBranches []string              `json:"protected_push_branches"`
	Ownership             []GovernanceOwnership `json:"ownership"`
	ProtectedKernel       []string              `json:"protected_kernel_paths"`
	Promotion             GovernancePromotion   `json:"promotion"`
}

type GovernanceContexts struct {
	Dev  []string `json:"dev"`
	Main []string `json:"main"`
}

type GuardianContexts struct {
	DevShadow    string `json:"dev_shadow"`
	MainRequired string `json:"main_required"`
}

type GovernanceOwnership struct {
	Branch string   `json:"branch"`
	Paths  []string `json:"paths"`
}

type GovernancePromotion struct {
	Source                   string   `json:"source"`
	Target                   string   `json:"target"`
	RequiredChecks           []string `json:"required_checks"`
	BranchProtectionRequired bool     `json:"branch_protection_required"`
}

// ReadGovernanceMatrix loads and validates the protected JSON contract.
func ReadGovernanceMatrix(filename string) (GovernanceMatrix, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return GovernanceMatrix{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var matrix GovernanceMatrix
	if err := decoder.Decode(&matrix); err != nil {
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
	if matrix.Mode != "ci_only" || matrix.CIAppID != 15368 || matrix.GuardianContexts.DevShadow != "CI guardian shadow" || matrix.GuardianContexts.MainRequired != "CI guardian" {
		return fmt.Errorf("governance mode must be ci_only")
	}
	if !sameStrings(matrix.ProofJobs, canonicalJobs()) {
		return fmt.Errorf("proof jobs do not match canonical CI jobs")
	}
	if !sameStrings(matrix.RequiredContexts.Dev, append(append([]string(nil), canonicalJobs()...), "CI guardian shadow")) || !sameStrings(matrix.RequiredContexts.Main, append(append([]string(nil), canonicalJobs()...), "CI guardian")) {
		return fmt.Errorf("required contexts do not match the dev/main guardian matrix")
	}
	if len(matrix.RequiredContexts.Dev) != len(canonicalJobs())+1 || len(matrix.RequiredContexts.Main) != len(matrix.RequiredContexts.Dev) || contains(matrix.RequiredContexts.Dev, "CI guardian") || !contains(matrix.RequiredContexts.Dev, "CI guardian shadow") {
		return fmt.Errorf("required context names are not unique and route-specific")
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

func validateKernelPaths(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("protected kernel paths are required")
	}
	for _, path := range paths {
		wildcard := strings.HasSuffix(path, "/**")
		base := strings.TrimSuffix(path, "/**")
		if path == "" || strings.ContainsAny(base, "*?[]") || strings.HasPrefix(path, "/") || (strings.Contains(path, "*") && !wildcard) {
			return fmt.Errorf("invalid protected kernel path %q", path)
		}
	}
	for _, required := range mandatoryKernelPaths() {
		if !contains(paths, required) {
			return fmt.Errorf("protected kernel paths cannot shrink below %q", required)
		}
	}
	return nil
}

func mandatoryKernelPaths() []string {
	return []string{
		".github/workflows/**",
		".github/ci-governance.json",
		".github/agent-scope-table.md",
		".github/branch-policy.md",
		".github/conformance-plan.md",
		"scripts/ci-proof/**",
		"scripts/ci-evidence/**",
		"scripts/verify/**",
		"internal/verify/**",
		"go.mod",
		"go.sum",
	}
}

func validatePromotion(promotion GovernancePromotion) error {
	if promotion.Source != "dev" || promotion.Target != "main" {
		return fmt.Errorf("promotion must be dev to main")
	}
	if !promotion.BranchProtectionRequired {
		return fmt.Errorf("promotion branch protection is required")
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
