package verify

func init() {
	branchScopeAllowlist["agent/gooo-closed-value"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/billing-package",
		"internal/meta/languageexampleexperiment",
		"internal/packageruntime/artifactemit",
		"internal/verify/scope_gooo_closed_value.go",
		"scripts/language-example-experiment",
	}
}
