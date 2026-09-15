package languageutility

import "strings"

func fixtureContract() Contract {
	return Contract{Schema: ContractSchema, ID: "gooo-language-utility-v1",
		Stages: append([]StageSpec(nil), CanonicalStages...),
		UseCases: []UseCaseSpec{
			{ID: "ci-plan-selection", Label: "CI plan"},
			{ID: "source-execution", Label: "Source execution"},
			{ID: "artifact-emission", Label: "Artifact emission"},
			{ID: "profiling", Label: "Profiling"},
			{ID: "debugging", Label: "Debugging"},
			{ID: "package-execution", Label: "Package execution"},
		}, Floors: Floors{ClosedCells: 39, CompleteUseCases: 4}}
}

func fixtureObservation(contract Contract) Observation {
	digest := "sha256:" + strings.Repeat("1", 64)
	value := Observation{Schema: ObservationSchema, ContractID: contract.ID,
		SubjectSHA: strings.Repeat("a", 40)}
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			value.Cells = append(value.Cells, CellObservation{
				UseCaseID: useCase.ID, StageID: stage.ID, State: StateClosed,
				Producer: "ci:" + useCase.ID, Step: "VERIFY_" + stage.ID,
				Reason: "EVIDENCE_ACCEPTED", EvidenceKey: useCase.ID + ".report",
				EvidencePath: "evidence/" + useCase.ID + ".json", EvidenceDigest: digest,
			})
		}
	}
	openCell(&value, "debugging", "DETERMINISTIC_REPLAY", "DEBUG_REPLAY_NOT_EXECUTED")
	openCell(&value, "debugging", "RESOURCE_OBSERVED", "DEBUG_RESOURCES_NOT_OBSERVED")
	openCell(&value, "package-execution", "RESOURCE_OBSERVED", "PACKAGE_RESOURCES_NOT_OBSERVED")
	return value
}

func openCell(value *Observation, useCase, stage, reason string) {
	for index := range value.Cells {
		cell := &value.Cells[index]
		if cell.UseCaseID == useCase && cell.StageID == stage {
			*cell = CellObservation{UseCaseID: useCase, StageID: stage, State: StateOpen,
				Producer: "ci:" + useCase, Step: "COLLECT_" + stage, Reason: reason}
			return
		}
	}
}
