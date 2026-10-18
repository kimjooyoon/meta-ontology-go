package verify

func init() {
	branchScopeAllowlist["agent/test-reuse-tool-origin-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"docs/language/language-product-priorities.ko.md",
		"internal/meta/publictestreuse/policy.gooo",
		"internal/meta/publictestreuse/receipt.go",
		"internal/meta/publictestreuse/receipt_inputs_test.go",
		"internal/meta/publictestreuse/tool_inputs.go",
		"internal/meta/publictestreuse/tool_inputs_test.go",
		"internal/toolchaincli/model.go",
		"internal/verify/scope_public_test_tool_origin.go",
		"scripts/self-improvement-public-test-reuse/assessment.go",
		"scripts/self-improvement-public-test-reuse/model.go",
		"scripts/self-improvement-public-test-reuse/receipt_inputs_test.go",
		"scripts/self-improvement-public-test-reuse/run.go",
		"scripts/self-improvement-public-test-reuse/tool_inputs.go",
		"scripts/self-improvement-public-test-reuse/tool_inputs_test.go",
	}
}
