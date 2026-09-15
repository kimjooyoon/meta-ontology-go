package verify

const recipeSourceBoundaryRepairV2Branch = "agent/repair-recipe-source-boundary-v2-20260916"

func init() {
	branchScopeAllowlist[recipeSourceBoundaryRepairV2Branch] = []string{
		"internal/meta/functionextractorrecipe/recipes.json",
		"internal/verify/scope_recipe_source_boundary_v2.go",
		".github/agent-scope-table.md",
		".github/ci-governance.json",
	}
}
