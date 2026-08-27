package languageutility

func summarize(contract Contract, observation Observation, cells []CellResult, issues int) (Summary, []UseCaseSummary) {
	summary := Summary{CellsTotal: len(cells), UseCasesTotal: len(contract.UseCases),
		ObservationIssues: issues, ClosedFloor: contract.Floors.ClosedCells,
		CompleteUseCaseFloor: contract.Floors.CompleteUseCases,
		RepositoryWrites: observation.RepositoryWrites}
	evidence := map[string]bool{}
	for _, cell := range cells {
		switch cell.State {
		case StateClosed:
			summary.ClosedCells++
			summary.ClaimsDischarged++
			evidence[cell.EvidenceKey] = true
		case StateOpen:
			summary.OpenCells++
			summary.ClaimsOpen++
		case StateUnknown:
			summary.UnknownCells++
			summary.ClaimsOpen++
		case StateRefuted:
			summary.RefutedCells++
			summary.ClaimsRefuted++
		}
	}
	summary.EvidenceArtifacts = len(evidence)
	summary.RemainingCells = summary.CellsTotal - summary.ClosedCells
	summary.ProgressBasisPoints = ratio(summary.ClosedCells, summary.CellsTotal)
	useCases := summarizeUseCases(contract, cells)
	for _, useCase := range useCases {
		if useCase.Complete {
			summary.CompleteUseCases++
		}
	}
	summary.ClosedDeltaFromFloor = summary.ClosedCells - summary.ClosedFloor
	summary.CompleteUseCaseFloorDelta = summary.CompleteUseCases - summary.CompleteUseCaseFloor
	summary.ObservationComplete = summary.UnknownCells == 0 && issues == 0 && len(cells) == 42
	summary.UtilityComplete = summary.ClosedCells == summary.CellsTotal
	summary.PromotionComplete = summary.UtilityComplete && summary.RefutedCells == 0 &&
		summary.ObservationComplete && summary.RepositoryWrites == 0
	return summary, useCases
}

func summarizeUseCases(contract Contract, cells []CellResult) []UseCaseSummary {
	result := make([]UseCaseSummary, 0, len(contract.UseCases))
	for _, spec := range contract.UseCases {
		value := UseCaseSummary{ID: spec.ID, TotalCells: len(contract.Stages)}
		for _, cell := range cells {
			if cell.UseCaseID == spec.ID && cell.State == StateClosed {
				value.ClosedCells++
			}
		}
		value.RemainingCells = value.TotalCells - value.ClosedCells
		value.Complete = value.RemainingCells == 0
		result = append(result, value)
	}
	return result
}

func ratio(value, total int) int {
	if total == 0 {
		return 0
	}
	return value * 10000 / total
}
