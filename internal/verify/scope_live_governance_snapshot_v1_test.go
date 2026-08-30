package verify

import "testing"

func TestLiveGovernanceSnapshotScope(t *testing.T) {
	paths, ok := BranchScope(liveGovernanceSnapshotBranch)
	if !ok || len(paths) != 13 {
		t.Fatalf("live governance branch registration: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/live-governance-snapshot.yml",
		"docs/external/live-governance-snapshot-v1.md",
		"examples/live-governance-snapshot/contract.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/governancesnapshot/evaluate.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_live_governance_snapshot_v1.go",
		"scripts/live-governance-snapshot/main.go",
	}
	if err := CheckPathScopeForBranch(allowed, liveGovernanceSnapshotBranch); err != nil {
		t.Fatalf("representative live governance paths rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, liveGovernanceSnapshotBranch); err == nil {
		t.Fatal("unrelated workflow accepted")
	}
}
