package verify

func init() {
	branchScopeAllowlist["agent/partial-reuse-evidence-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/publicpartialreuse/evaluator.go",
		"internal/meta/publicpartialreuse/evaluator_receipt_test.go",
		"internal/meta/publicpartialreuse/receipt.go",
		"internal/meta/publicpartialreuse/receipt_failure_test.go",
		"internal/meta/publicpartialreuse/receipt_fixture_test.go",
		"internal/verify/scope_partial_reuse_evidence_20260911.go",
		"scripts/self-improvement-public-partial-reuse/run.go",
	}
}
