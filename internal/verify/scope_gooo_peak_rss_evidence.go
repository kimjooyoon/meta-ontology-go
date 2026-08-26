package verify

func init() {
	branchScopeAllowlist["agent/gooo-peak-rss-evidence"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/symbolic-invocation-usecase",
		"internal/meta/symbolicinvocationusecase",
		"internal/verify/scope_gooo_peak_rss_evidence.go",
		"scripts/symbolic-invocation-usecase",
	}
}
