package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const GovernanceSchemaVersion = "gooo/ci-governance/v2"

const foundationCorrectionDevBranch = "agent/foundation-correction-dev-20260831"

func init() {
	branchScopeAllowlist[foundationCorrectionDevBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/foundation-bootstrap-dev-sync.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/governance_part01.go",
		"scripts/ci-proof/foundation_bootstrap.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/route_test.js",
	}
}

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
