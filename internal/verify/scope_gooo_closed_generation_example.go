package verify

func init() {
	paths := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/symbolic-invocation-schema",
		"examples/symbolic-invocation-usecase",
		"internal/meta/symbolicinvocationusecase",
		"internal/packageruntime/artifactemit",
		"internal/verify/scope_gooo_closed_generation_example.go",
		"scripts/symbolic-invocation-schema",
		"scripts/symbolic-invocation-usecase",
	}
	branchScopeAllowlist["agent/gooo-closed-generation-example"] = paths
	branchScopeAllowlist["agent/gooo-reader-resolution-projection"] = paths
}
