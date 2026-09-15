package verify

func init() {
	branchScopeAllowlist["agent/v24-candidate-authorization-bridge-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-candidate-authorization.yml",
		".github/workflows/transformation-effect.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement/AUTHORIZATION.md",
		"examples/self-improvement/authorization.gooo",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/selfimprovementcandidate/authorization.go",
		"internal/meta/selfimprovementcandidate/authorization_test.go",
		"internal/verify/scope_gooo_candidate_authorization_v24.go",
		"scripts/self-improvement-candidate-authorization",
	}
}
