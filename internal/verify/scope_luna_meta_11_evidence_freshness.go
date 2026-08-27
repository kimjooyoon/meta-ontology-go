package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-11-evidence-freshness"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/evidence-freshness.yml",
		"cmd/evidence-freshness",
		"cmd/evidence-freshness-decider",
		"docs/language/evidence-freshness.md",
		"examples/evidence-freshness",
		"internal/meta/evidencefreshness",
		"internal/syntax",
		"internal/verify/scope_luna_meta_11_evidence_freshness.go",
		"scripts/evidence-freshness",
	}
}
