package selfimprovementloop

func bindingMap(bindings []ActivityBinding) map[string]ActivityBinding {
	result := make(map[string]ActivityBinding, len(bindings))
	for _, binding := range bindings {
		result[binding.Cell] = binding
	}
	return result
}

func makeCellResult(bindings map[string]ActivityBinding, cell string, evaluation cellEvaluation) CellResult {
	binding := bindings[cell]
	return CellResult{
		Cell: cell, Activity: binding.Activity, ActivityID: binding.ActivityID,
		Decision: evaluation.Decision, Reason: evaluation.Reason, Unknown: evaluation.Unknown,
	}
}

func summarizeResults(results []CellResult) ([]string, []UnknownState) {
	decisions := make([]string, 0, len(results))
	unknowns := make([]UnknownState, 0)
	for _, result := range results {
		decisions = append(decisions, result.Decision)
		if result.Unknown != nil {
			unknowns = append(unknowns, *result.Unknown)
		}
	}
	return decisions, unknowns
}

func unknownCellNames(results []CellResult) []string {
	result := make([]string, 0)
	for _, cell := range results {
		if cell.Decision == DecisionUnknown {
			result = append(result, cell.Cell)
		}
	}
	return result
}
