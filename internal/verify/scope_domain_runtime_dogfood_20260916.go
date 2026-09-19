package verify

func init() {
	branchScopeAllowlist["agent/domain-runtime-dogfood-20260916"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/domain-observation.yml",
		"examples/domain-observation/README.md",
		"examples/domain-observation/definition.gooo",
		"examples/domain-observation/input.json",
		"examples/domain-observation/main.gooo",
		"examples/domain-observation/observe.sh",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator_migration_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v31.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_domain_runtime_dogfood_20260916.go",
	}
}
