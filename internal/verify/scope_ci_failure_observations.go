package verify

func init() {
	branchScopeAllowlist["agent/ci-failure-observations-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"docs/ci-work-domain.ko.md",
		"internal/verify/scope_ci_failure_observations.go",
		"scripts/ci-proof/failure_capture_test.js",
		"scripts/ci-proof/refs_test.js",
	}
}
