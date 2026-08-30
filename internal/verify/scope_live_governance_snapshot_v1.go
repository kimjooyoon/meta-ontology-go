package verify

const liveGovernanceSnapshotBranch = "agent/live-governance-snapshot-v1"

func init() {
	branchScopeAllowlist[liveGovernanceSnapshotBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/causal-ci-selection.yml",
		".github/workflows/live-governance-snapshot.yml",
		"docs/external/live-governance-snapshot-v1.md",
		"examples/live-governance-snapshot",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/governancesnapshot",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_live_governance_snapshot_v1.go",
		"internal/verify/scope_live_governance_snapshot_v1_test.go",
		"scripts/live-governance-snapshot",
	}
}
