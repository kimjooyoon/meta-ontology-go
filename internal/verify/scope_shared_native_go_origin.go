package verify

func init() {
	branchScopeAllowlist["agent/shared-native-go-origin-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/self-improvement-partial-reuse/README.md",
		"internal/meta/publicpartialreuse/receipt.go",
		"internal/meta/publicpartialreuse/resource_origin_test.go",
		"internal/meta/publictestreuse/receipt.go",
		"internal/meta/publictestreuse/tool_inputs.go",
		"internal/meta/publictestreuse/tool_inputs_test.go",
		"internal/toolchaincli/go_command.go",
		"internal/toolchaincli/go_command_test.go",
		"internal/toolchaincli/go_inputs.go",
		"internal/verify/scope_shared_native_go_origin.go",
		"scripts/self-improvement-public-partial-reuse/model.go",
		"scripts/self-improvement-public-partial-reuse/native_timing_test.go",
		"scripts/self-improvement-public-partial-reuse/run.go",
		"scripts/self-improvement-public-test-reuse/assessment.go",
		"scripts/self-improvement-public-test-reuse/model.go",
		"scripts/self-improvement-public-test-reuse/run.go",
		"scripts/self-improvement-public-test-reuse/tool_inputs.go",
		"scripts/self-improvement-public-test-reuse/tool_inputs_test.go",
	}
}
