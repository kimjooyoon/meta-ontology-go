package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-11-evidence-freshness"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/evidence-freshness",
		"cmd/evidence-freshness-decider",
		"docs/language/evidence-freshness.md",
		"examples/evidence-freshness",
		"internal/meta/evidencefreshness",
		"internal/verify/scope_luna_meta_11_evidence_freshness.go",
		"scripts/evidence-freshness",
	}
}
