package verify

const (
	ciWorkflowBranch             = "agent/ci-workflow"
	recipeSourceBoundaryBranch   = "agent/repair-recipe-source-boundary-20260916"
)

func init() {
	branchScopeAllowlist[ciWorkflowBranch] = []string{
		".github/**",
		"internal/verify/**",
		"scripts/**",
	}
	branchScopeAllowlist[recipeSourceBoundaryBranch] = []string{
		"internal/meta/functionextractorrecipe/recipes.json",
	}
}
