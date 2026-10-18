package verify

func init() {
	branchScopeAllowlist["agent/gooo-nullary-helpers-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/constant-helpers/README.md",
		"examples/constant-helpers/cases.json",
		"examples/constant-helpers/gooo.workspace.json",
		"examples/constant-helpers/main.gooo",
		"internal/bodycodegen/pure_calls.go",
		"internal/bodycodegen/pure_calls_nullary_test.go",
		"internal/packageruntime/workspaceexecution/body_calls_nullary_test.go",
		"internal/verify/scope_gooo_nullary_helpers.go",
	}
}
