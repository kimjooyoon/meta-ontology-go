package verify

func init() {
	branchScopeAllowlist["agent/gooo-package-exec-tiny-model-20261005"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_execute.go",
		"cmd/gooo/package_execute_test.go",
		"cmd/gooo/package_execute_tiny_model_options_test.go",
		"cmd/gooo/package_execute_tiny_model_test.go",
		"cmd/gooo/templates/library/README.md",
		"cmd/gooo/templates/library/body-fill-plans.json",
		"docs/language/project-starters.md",
		"docs/language/workspace-manifest.md",
		"internal/packageruntime/workspaceexecution/execute.go",
		"internal/verify/scope_gooo_package_exec_tiny_model.go",
	}
}
