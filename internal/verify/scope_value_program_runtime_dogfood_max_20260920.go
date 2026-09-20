package verify

func init() {
	branchScopeAllowlist["agent/value-program-runtime-dogfood-max-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-value-witness.yml",
		"examples/language-value-witness/README.md",
		"examples/language-value-witness/max.gooo",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/language-semantic-model/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v40.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator_migration_test.go",
		"internal/verify/scope_value_program_runtime_dogfood_max_20260920.go",
		"internal/valueexecution/evaluate_test.go",
		"internal/valueexecution/measurement.go",
		"internal/valueexecution/operation_spec_validate.go",
		"internal/valueexecution/registry.go",
		"internal/valueexecution/validate.go",
	}
}
