package verify

func init() {
	branchScopeAllowlist["agent/gooo-candidate-lineage-fix"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		"internal/verify/scope_gooo_candidate_lineage_fix.go",
	}
}
