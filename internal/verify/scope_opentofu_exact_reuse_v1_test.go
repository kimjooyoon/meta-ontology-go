package verify

import "testing"

func TestOpenTofuExactReuseScope(t *testing.T) {
	paths, ok := BranchScope(opentofuExactReuseBranch)
	if !ok || len(paths) != 6 {
		t.Fatalf("exact reuse branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/opentofuobservation/validate.go",
		"scripts/opentofu-observation/main.sh",
	}
	if err := CheckPathScopeForBranch(allowed, opentofuExactReuseBranch); err != nil {
		t.Fatalf("representative exact reuse paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"internal/meta/opentofuobservation/validate_other.go"}, opentofuExactReuseBranch); err == nil {
		t.Fatal("unregistered OpenTofu package path was accepted")
	}
	if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, opentofuExactReuseBranch); err == nil {
		t.Fatal("unrelated workflow path was accepted")
	}
}
