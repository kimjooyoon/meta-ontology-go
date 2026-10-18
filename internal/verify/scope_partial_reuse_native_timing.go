package verify

func init() {
	branchScopeAllowlist["agent/partial-reuse-native-timing-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/self-improvement-partial-reuse/README.md",
		"internal/meta/publicpartialreuse/receipt.go",
		"internal/meta/publicpartialreuse/resource_origin_test.go",
		"internal/toolchaincli/model.go",
		"internal/verify/scope_partial_reuse_native_timing.go",
		"scripts/self-improvement-public-partial-reuse/native_timing_test.go",
		"scripts/self-improvement-public-partial-reuse/run.go",
	}
}
