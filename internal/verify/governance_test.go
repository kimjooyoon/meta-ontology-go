package verify

import (
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
}

func TestGovernanceMatrixRejectsWildcardOwnership(t *testing.T) {
	matrix := GovernanceMatrix{Schema: GovernanceSchemaVersion, Ownership: []GovernanceOwnership{{Branch: "agent/*", Paths: []string{"internal/verify"}}}}
	if err := validateGovernanceOwnership(matrix.Ownership); err == nil {
		t.Fatal("wildcard ownership was accepted")
	}
}
