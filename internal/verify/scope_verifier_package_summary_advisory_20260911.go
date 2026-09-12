package verify

func init() {
	branchScopeAllowlist["agent/verifier-package-summary-advisory-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"internal/verify/scope_verifier_package_summary_advisory_20260911.go",
		"scripts/meta-execution/observation_journal.go",
		"scripts/meta-execution/run.go",
		"scripts/meta-execution/trace.go",
		"scripts/meta-execution/verifier_package_summary.go",
		"scripts/meta-execution/verifier_package_summary_test.go",
	}
}
