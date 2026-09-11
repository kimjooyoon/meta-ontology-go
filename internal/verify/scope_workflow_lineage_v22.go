package verify

const workflowLineageV22Branch = "agent/resolve-workflow-run-lineage-v22"

func init() {
	branchScopeAllowlist[workflowLineageV22Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-effort-observation.yml",
		"examples/ci-effort-observation",
		"internal/meta/publicworkflowlineage",
		"internal/verify/scope_workflow_lineage_v22.go",
		"scripts/ci-effort-observation-lineage",
	}
}
