package verify

func init() {
	branchScopeAllowlist["agent/gooo-exact-subject-transport"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		".github/workflows/self-improvement-language-observation.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement/TRANSPORT.md",
		"examples/self-improvement/transport.gooo",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_exact_subject_transport.go",
		"scripts/self-improvement-transport",
	}
}
