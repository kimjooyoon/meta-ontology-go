package verify

func init() {
	branchScopeAllowlist["agent/ci-concurrency-lane"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"examples/ci-artifact-lineage/usecases.json",
		"examples/ci-concurrency-lane",
		"internal/verify/scope_ci_concurrency_lane.go",
		"scripts/ci-proof/artifacts.js",
		"scripts/ci-proof/artifacts_test.js",
		"scripts/ci-proof/concurrency.js",
		"scripts/ci-proof/concurrency_test.js",
	}
}
