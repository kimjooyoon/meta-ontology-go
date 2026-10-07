package bodycodegen

import "fmt"

func (p preparedBody) selectedResult(source []byte, choice bodyRouteChoice) (Result, error) {
	result := p.base
	if choice.selected != preserveRoute {
		candidate, err := p.generate(choice.selected)
		if err == nil {
			result = candidate
		} else {
			choice.selected, choice.receipt.Selected = preserveRoute, preserveRoute
			choice.receipt.Mode, choice.receipt.Provider = "deterministic_fallback", "deterministic"
			choice.receipt.FallbackReason = "SELECTED_ROUTE_LOWERING_FAILED"
			choice.selection.Method = "deterministic_fallback_after_lowering_failure"
		}
	}
	choice.selection.FinalRoute = choice.selected
	replay, err := p.generate(choice.selected)
	if err != nil {
		return Result{}, fmt.Errorf("replay selected code generation route: %w", err)
	}
	r, next := result.report, replay.report
	if r.SourceConstructs != next.SourceConstructs || r.LoweredConstructs != next.LoweredConstructs ||
		r.SourceSemanticUnits != next.SourceSemanticUnits || r.LoweredSemanticUnits != next.LoweredSemanticUnits {
		return Result{}, fmt.Errorf("internal error: replay construct count changed")
	}
	result.report.SourceDigest, result.report.ProgramDigest = digest(source), digest([]byte(p.activity.ValueProgram))
	result.report.GeneratedDigest, result.report.ReplayDigest = digest(result.source), digest(replay.source)
	result.report.Route, result.report.RouteDecision, result.report.RouteSelection = choice.selected, choice.receipt, choice.selection
	result.report.RouteDecisionLatencyMS, result.report.CandidateRoutes = choice.latencyMS, choice.ids
	result.report.DeterministicReplay = result.report.GeneratedDigest == result.report.ReplayDigest
	result.report.RepositoryWrites = 0
	result.report.UnsupportedConstructs = "calls, loops, imports, external effects, optional-field synthesis and repeated record fields"
	populateCompletenessReceipt(&result.report, "")
	return Result{Report: result.report, Source: string(result.source)}, nil
}
