package verify

const proposalPredecessorRouteIdentity20260903Branch = "agent/proposal-predecessor-route-identity-20260903"

func init() {
	branchScopeAllowlist[proposalPredecessorRouteIdentity20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/proposal-promotion/main.go",
		"cmd/language-readiness-witness/proposal-promotion/run.go",
		"docs/language/proposal-synthesis-continuity.md",
		"internal/meta/languagereadiness/proposalpromotion/coordinates.go",
		"internal/meta/languagereadiness/proposalpromotion/source.go",
		"internal/meta/languagereadiness/proposalpromotion/source_build.go",
		"internal/meta/languagereadiness/proposalpromotion/validate.go",
		"internal/meta/languagereadiness/proposalpromotion/validate_test.go",
		"internal/meta/languagereadiness/proposalpromotion/views.go",
		"internal/meta/metricstrategy/proposalpredecessor/candidate.go",
		"internal/meta/metricstrategy/proposalpredecessor/collect.go",
		"internal/meta/metricstrategy/proposalpredecessor/failure.go",
		"internal/meta/metricstrategy/proposalpredecessor/github.go",
		"internal/meta/metricstrategy/proposalpredecessor/indicators.go",
		"internal/meta/metricstrategy/proposalpredecessor/proofs.go",
		"internal/meta/metricstrategy/proposalpredecessor/resolution.go",
		"internal/meta/metricstrategy/proposalpredecessor/resolution_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/route_identity_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/select.go",
		"internal/meta/metricstrategy/proposalpredecessor/select_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/synthesis_job_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/types.go",
		"internal/meta/metricstrategy/proposalpredecessor/validate.go",
		"internal/verify/scope_proposal_predecessor_route_identity_20260903.go",
		"scripts/metric-strategy/main.go",
		"scripts/metric-strategy/predecessor.go",
	}
}
