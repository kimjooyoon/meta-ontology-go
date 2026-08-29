package verify

const transformationExecutorBindingBranch = "agent/transformation-executor-binding-v1"

func init() {
	branchScopeAllowlist[transformationExecutorBindingBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/generation/registry.go",
		"internal/meta/transformationeffect",
		"internal/verify/scope_transformation_executor_binding_v1.go",
	}
}
