package languageutility

import "fmt"

func Evaluate(contract Contract, observation Observation) (Report, error) {
	if err := ValidateContract(contract); err != nil {
		return Report{}, err
	}
	if observation.Schema != ObservationSchema || observation.ContractID != contract.ID ||
		len(observation.SubjectSHA) != 40 {
		return Report{}, fmt.Errorf("language utility observation identity is invalid")
	}
	index := indexObservation(contract, observation)
	cells := make([]CellResult, 0, 42)
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			key := cellKey(useCase.ID, stage.ID)
			observed, exists := index.cells[key]
			if !exists {
				cells = append(cells, classifyCell(useCase, stage, nil, false))
				continue
			}
			cells = append(cells, classifyCell(useCase, stage, &observed, index.duplicates[key]))
		}
	}
	summary, useCases := summarize(contract, observation, cells, index.issues)
	contractDigest, err := digestJSON(contract)
	if err != nil {
		return Report{}, err
	}
	observationDigest, err := digestJSON(observation)
	if err != nil {
		return Report{}, err
	}
	program, err := GenerateProgram(contract)
	if err != nil {
		return Report{}, err
	}
	report := Report{Schema: ReportSchema, ContractID: contract.ID, SubjectSHA: observation.SubjectSHA,
		Summary: summary, UseCases: useCases, Cells: cells, NotClaimed: contract.NotClaimed,
		ContractDigest: contractDigest, ObservationDigest: observationDigest, ProgramDigest: digestBytes([]byte(program))}
	report.Decision, report.Resolution, report.Reason = decide(summary)
	report.Indicators = buildIndicators(summary)
	report.Proofs = buildProofs(cells, observationDigest)
	report.Digest, err = digestJSON(report)
	return report, err
}

func decide(summary Summary) (string, string, string) {
	if summary.UnknownCells > 0 || summary.ObservationIssues > 0 {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "UTILITY_EVIDENCE_UNKNOWN"
	}
	if summary.RefutedCells > 0 {
		return "FAIL_CLOSED", "EXACT", "UTILITY_CLAIM_REFUTED"
	}
	if summary.RepositoryWrites != 0 {
		return "FAIL_CLOSED", "EXACT", "UTILITY_OBSERVER_MUTATED_REPOSITORY"
	}
	if summary.ClosedDeltaFromFloor < 0 || summary.CompleteUseCaseFloorDelta < 0 {
		return "FAIL_CLOSED", "EXACT", "UTILITY_FLOOR_REGRESSION"
	}
	if summary.UtilityComplete {
		return "UTILITY_COMPLETE", "EXACT", "ALL_UTILITY_CELLS_CLOSED"
	}
	return "PROGRESS_OBSERVED", "EXACT", "UTILITY_GAPS_REMAIN"
}
