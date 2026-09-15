package verify

const recipeSourceBoundaryRepairBranch = "agent/repair-recipe-source-boundary-20260916"

func init() {
	branchScopeAllowlist[recipeSourceBoundaryRepairBranch] = []string{
		"internal/meta/functionextractorrecipe/recipes.json",
		"internal/verify/scope_recipe_source_boundary.go",
	}
}
