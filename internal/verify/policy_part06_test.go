package verify

import (
	"strings"
	"testing"
)

func TestFollowUpScopeBoundariesAndUnknowns(t *testing.T) {
	cases := []struct {
		branch string
		path   string
	}{
		{"agent/codegen-hypotheses", "docs/research/codegen.md"},
		{"agent/codegen-fixture-adapter", "docs/research/codegen-experiments.md"},
		{"agent/bidir-followup", "internal/semantic/graph.go"},
		{"agent/freshness-research", "internal/research/other/contract.go"},
		{"agent/integration-governance", "docs/research/integration-promotion.md"},
		{"agent/integration-governance-followup", "docs/governance/other.md"},
		{"agent/lsp-contracts", "docs/research/grammar.md"},
		{"agent/testing-research-contracts", "docs/research/security.md"},
		{"agent/ci-ownership-audit", "internal/semantic/graph.go"},
		{"agent/ci-ownership-audit-current", "internal/semantic/graph.go"},
		{"agent/ci-ownership-audit-current2", "internal/semantic/graph.go"},
		{"agent/ci-alias-refresh", "internal/semantic/graph.go"},
		{"agent/ci-generator-current7", "internal/semantic/graph.go"},
		{"agent/cli-check", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current2", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current5", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current7", "internal/semantic/graph.go"},
	}
	for _, test := range cases {
		if err := CheckPathScopeForBranch([]string{test.path}, test.branch); err == nil {
			t.Errorf("%s incorrectly allowed %s", test.branch, test.path)
		}
	}
	for _, branch := range []string{"agent/unknown-followup", "agent/freshness-unknown"} {
		if err := CheckPathScopeForBranch([]string{"docs/research/unknown.md"}, branch); err == nil {
			t.Errorf("unknown branch %s was not rejected", branch)
		}
	}
}
func TestBidirFollowUpAliasIsExact(t *testing.T) {
	paths, ok := BranchScope("agent/bidir-followup")
	if !ok || len(paths) != 1 || paths[0] != "internal/bidir" {
		t.Fatalf("unexpected bidir-followup ownership: %#v", paths)
	}
	if err := CheckPathScopeForBranch([]string{"internal/bidir/hosting_contract.go"}, "agent/bidir-followup"); err != nil {
		t.Fatal(err)
	}
}
func TestAllowlistKeysAndSelfHostingPathsAreUnique(t *testing.T) {
	branches := ConfiguredBranches()
	if len(branches) != len(sortedUnique(branches)) {
		t.Fatalf("duplicate branch keys detected: %#v", branches)
	}
	for _, branch := range branches {
		if strings.ContainsAny(branch, "*?") {
			t.Fatalf("wildcard branch key weakens fail-closed policy: %q", branch)
		}
	}
	paths, ok := BranchScope("agent/self-hosting-bootstrap")
	if !ok || len(paths) != len(sortedUnique(paths)) {
		t.Fatalf("duplicate self-hosting alias paths detected: %#v", paths)
	}
}
