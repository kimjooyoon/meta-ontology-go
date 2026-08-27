package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-23-nonmonotonic-refutation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/nonmonotonic-refutation.yml",
		"cmd/nonmonotonic-refutation-consumer",
		"cmd/nonmonotonic-refutation-producer",
		"examples/nonmonotonic-refutation",
		"internal/meta/nonmonotonicrefutation",
		"internal/meta/nonmonotonicrefutationoracle",
		"internal/verify/scope_luna_meta_23.go",
	}
}
