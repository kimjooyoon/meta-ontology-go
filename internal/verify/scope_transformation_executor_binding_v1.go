package verify

const transformationExecutorBindingBranch = "agent/transformation-executor-binding-v1"

func init() {
	branchScopeAllowlist[transformationExecutorBindingBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"internal/meta/generation/operation_observation.go",
		"internal/meta/generation/registry.go",
		"internal/meta/generation/receipt_types_part02.go",
		"internal/meta/transformationeffect",
		"internal/meta/transformationeffectverification",
		"internal/verify/scope_transformation_executor_binding_v1.go",
		"scripts/meta-execution/operations.go",
		"scripts/transformation-effect-verify",
	}
}
