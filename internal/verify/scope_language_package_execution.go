package verify

func init() {
	branchScopeAllowlist["agent/language-multi-file-execution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-package-execution.yml",
		"cmd/gooo/run_package_source.go",
		"cmd/gooo/run_package_source_test.go",
		"cmd/gooo/run_source_part01.go",
		"cmd/gooo/run_source_part02.go",
		"cmd/language-package-execution-witness",
		"docs/language/language-package-execution.md",
		"examples/billing-package",
		"examples/language-delivery-scorecard/contract.json",
		"examples/language-package-execution",
		"examples/user-journey-scorecard/contract.json",
		"internal/meta/languagedelivery",
		"internal/meta/languagepackageexecution",
		"internal/meta/userjourneyscorecard",
		"internal/language/packageexecution",
		"internal/verify/scope_language_package_execution.go",
		"scripts/language-package-execution",
	}
}
