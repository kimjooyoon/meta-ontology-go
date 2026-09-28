package verify

func init() {
	branchScopeAllowlist["agent/meta-cost-publication-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-cost-report.yml",
		"internal/verify/scope_meta_cost_publication_20260907.go",
		"scripts/meta-cost-report/capture.mjs",
	}
}
