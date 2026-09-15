package verify

func init() {
	branchScopeAllowlist["agent/v26-separate-execution-grant-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-execution-grant.yml",
		".github/workflows/transformation-effect.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement-execution-grant",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/selfimprovementexecutiongrant",
		"internal/verify/scope_gooo_self_improvement_execution_grant_v26.go",
		"scripts/self-improvement-execution-grant",
	}
}
