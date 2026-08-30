package verify

const entityFieldsV1Branch = "agent/entity-fields-v1"

func init() {
	branchScopeAllowlist[entityFieldsV1Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/entity-fields-v1.yml",
		"cmd/entity-fields-witness",
		"docs/language/language-syntax-roundtrip.md",
		"examples/entity-fields-v1",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/bidir/entity_fields_public.go",
		"internal/bidir/field_span_validation.go",
		"internal/entityfieldsv1",
		"internal/generator/generator_part01.go",
		"internal/lsp/adapter_part01.go",
		"internal/lsp/adapter_part04.go",
		"internal/lsp/entity_fields_parser.go",
		"internal/lsp/parser_part02.go",
		"internal/meta/entityfields",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/evaluate.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/languagesyntax/replay/execute.go",
		"internal/meta/languagereadiness/languagesyntax/replay/semantic.go",
		"internal/syntax/entity_fields_support.go",
		"internal/verify/scope_entity_fields_v1.go",
		"internal/verify/scope_entity_fields_v1_test.go",
	}
}
