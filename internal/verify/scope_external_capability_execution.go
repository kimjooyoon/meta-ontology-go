package verify

func init() {
	paths := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/external-capability-execution-witness",
		"examples/external-capability-execution",
		"internal/meta/externalcapabilityexecution",
		"internal/verify/scope_external_capability_execution.go",
	}
	branches := []string{"agent/external-capability-execution", "agent/external-assurance-eligibility",
		"agent/gooo-capability-boundary"}
	for _, branch := range branches {
		branchPaths := append([]string(nil), paths...)
		if branch == "agent/gooo-capability-boundary" {
			branchPaths = append(branchPaths, "examples/language-semantic-model/corpus.json",
				"examples/language-syntax-roundtrip/corpus.json",
				"examples/toolchain-conformance/corpus.json",
				"examples/vertical-slice-closure-shadow/README.md",
				"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
				"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
				"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
				"internal/meta/languagereadiness/languagesemantic/model.go",
				"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
				"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
				"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
				"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
				"internal/meta/languagereadiness/languagesyntax/model.go",
				"internal/meta/languagereadiness/languagesyntax/registry.go",
				"internal/meta/languagereadiness/toolchainconformance/contract.go",
				"internal/meta/languagereadiness/toolchainconformance/corpus.go",
				"internal/meta/languagereadiness/toolchainconformance/evaluate_test.go")
		}
		branchScopeAllowlist[branch] = branchPaths
	}
}
