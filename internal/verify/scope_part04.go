package verify

func init() {
	branchScopeAllowlist["agent/gooo-record-relation-triples-20261007"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"docs/language-direction.ko.md",
		"docs/language/body-codegen.md",
		"docs/source-assembly.md",
		"examples/body-codegen/source-ir-fill-record-relation-triples.gooo.fixture",
		"internal/assemblyspec/spec.go",
		"internal/assemblyspec/spec_test.go",
		"internal/bodycodegen/body_fill_record_test.go",
		"internal/bodycodegen/source_record_fill_candidates.go",
		"internal/verify/scope_part04.go",
	}
}
