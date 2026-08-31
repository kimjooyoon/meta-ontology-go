package verify

func init() {
	branchScopeAllowlist["agent/gooo-nonexecuting-candidate"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement/CANDIDATE.md",
		"examples/self-improvement/candidate.gooo",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/selfimprovementcandidate",
		"internal/verify/scope_gooo_nonexecuting_candidate.go",
		"scripts/self-improvement-candidate",
	}
}
