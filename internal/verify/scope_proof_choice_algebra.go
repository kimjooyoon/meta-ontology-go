package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-07-proof-choice-algebra"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/proof-choice-algebra.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/proof-choice-algebra",
		"examples/toolchain-conformance/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/toolchainconformance/corpus.go",
		"internal/meta/languagereadiness/toolchainconformance/evaluate_test.go",
		"internal/meta/proofchoicealgebra",
		"internal/meta/proofchoicejudge",
		"internal/verify/scope_proof_choice_algebra.go",
		"scripts/proof-choice-algebra",
	}
}
