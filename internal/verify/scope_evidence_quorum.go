package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-28-evidence-quorum"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/evidence-quorum-witness",
		"examples/evidence-quorum",
		"internal/meta/evidencequorum",
		"internal/meta/languageconcept",
		"internal/verify/scope_evidence_quorum.go",
		"scripts/evidence-quorum",
	}
}
