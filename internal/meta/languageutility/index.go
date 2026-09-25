package languageutility

type observationIndex struct {
	cells      map[string]CellObservation
	duplicates map[string]bool
	issues     int
}

func indexObservation(contract Contract, observation Observation) observationIndex {
	expected := map[string]bool{}
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			expected[cellKey(useCase.ID, stage.ID)] = true
		}
	}
	result := observationIndex{
		cells: map[string]CellObservation{}, duplicates: map[string]bool{},
	}
	for _, cell := range observation.Cells {
		key := cellKey(cell.UseCaseID, cell.StageID)
		if !expected[key] {
			result.issues++
			continue
		}
		if _, exists := result.cells[key]; exists {
			result.duplicates[key] = true
			continue
		}
		result.cells[key] = cell
	}
	return result
}

func cellKey(useCase, stage string) string {
	return useCase + "\x00" + stage
}
