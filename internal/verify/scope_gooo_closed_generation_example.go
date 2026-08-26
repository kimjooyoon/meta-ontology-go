package verify

func init() {
	branchScopeAllowlist["agent/gooo-closed-generation-example"] = []string{
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
}
