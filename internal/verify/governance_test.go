package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGovernanceMatrixMatchesExecutablePolicy(t *testing.T) {
	filename := filepath.Join("..", "..", ".github", "ci-governance.json")
	matrix, err := ReadGovernanceMatrix(filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix.Ownership) != len(branchScopeAllowlist) {
		t.Fatalf("matrix ownership rows=%d executable rows=%d", len(matrix.Ownership), len(branchScopeAllowlist))
	}
	if matrix.Mode != "ci_only" {
		t.Fatalf("governance mode=%q, want ci_only", matrix.Mode)
	}
}

func TestGovernanceMatrixRejectsHumanReviewPolicyFields(t *testing.T) {
	filename := filepath.Join("..", "..", ".github", "ci-governance.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	promotion := document["promotion"].(map[string]any)
	promotion["required_pull_request_reviews"] = 1
	mutated, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	mutatedPath := filepath.Join(t.TempDir(), "ci-governance.json")
	if err := os.WriteFile(mutatedPath, mutated, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGovernanceMatrix(mutatedPath); err == nil {
		t.Fatal("human review policy was silently accepted in ci_only governance")
	}
}

func TestGovernanceMatrixRejectsNonCIMode(t *testing.T) {
	matrix, err := ReadGovernanceMatrix(filepath.Join("..", "..", ".github", "ci-governance.json"))
	if err != nil {
		t.Fatal(err)
	}
	matrix.Mode = "human_review"
	if err := ValidateGovernanceMatrix(matrix); err == nil {
		t.Fatal("non-ci governance mode was accepted")
	}
}

func TestGovernanceMatrixRejectsWildcardOwnership(t *testing.T) {
	matrix := GovernanceMatrix{Schema: GovernanceSchemaVersion, Ownership: []GovernanceOwnership{{Branch: "agent/*", Paths: []string{"internal/verify"}}}}
	if err := validateGovernanceOwnership(matrix.Ownership); err == nil {
		t.Fatal("wildcard ownership was accepted")
	}
}
