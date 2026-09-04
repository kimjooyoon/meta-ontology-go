package verify

func init() {
	branchScopeAllowlist["agent/v25-candidate-execution-contract-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-execution-contract.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement-execution-contract/**",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/selfimprovementexecutioncontract/**",
		"internal/verify/scope_gooo_self_improvement_execution_contract_v25.go",
		"scripts/self-improvement-execution-contract/**",
	}
}
