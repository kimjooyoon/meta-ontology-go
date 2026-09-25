package verify

const liveGovernanceSnapshotBranch = "agent/live-governance-snapshot-v1"

func init() {
	branchScopeAllowlist[liveGovernanceSnapshotBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/causal-ci-selection.yml",
		".github/workflows/live-governance-snapshot.yml",
		".github/workflows/transformation-effect.yml",
		"docs/language/language-syntax-roundtrip.md",
		"docs/external/live-governance-snapshot-v1.md",
		"examples/live-governance-snapshot",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/governancesnapshot",
		"internal/meta/languageassurance/verticalsliceclosureshadow/artifact_link_values.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/artifact_values.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/fixture_language_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/fixture_toolchain_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/model_summary.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/finish.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/report.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_live_governance_snapshot_v1.go",
		"internal/verify/scope_live_governance_snapshot_v1_test.go",
		"scripts/live-governance-snapshot",
	}
}
