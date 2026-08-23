package verify

func init() {
	branchScopeAllowlist["agent/source-splitter-conformance-governance"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_splitter_conformance.go",
	}
	branchScopeAllowlist["agent/source-splitter-conformance-fail-closed"] = []string{
		"internal/meta/transformationeffect",
	}
	branchScopeAllowlist["agent/source-splitter-conformance-contract"] = []string{
		"docs/language/source-splitter-operation-conformance-v1.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/source-splitter-conformance",
	}
	branchScopeAllowlist["agent/source-splitter-conformance-evaluator"] = []string{
		"internal/meta/operationconformance",
		"scripts/source-splitter",
	}
	branchScopeAllowlist["agent/source-splitter-conformance-adapter"] = []string{
		"internal/meta/generation/registry.go",
		"internal/meta/transformationeffect",
	}
	branchScopeAllowlist["agent/source-splitter-conformance-ci"] = []string{
		".github/workflows/transformation-effect.yml",
		"scripts/source-splitter-conformance-ci",
	}
}
