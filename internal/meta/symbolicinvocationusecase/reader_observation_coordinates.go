package symbolicinvocationusecase

func readerObservationCoordinates(indicators []ReaderObservationIndicator) ReaderObservationCoordinates {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	return ReaderObservationCoordinates{
		Satisfied:   satisfied,
		Total:       len(indicators),
		BasisPoints: satisfied * 10000 / len(indicators),
	}
}

func readerObservationClasses(indicators []ReaderObservationIndicator) []ReaderObservationClassCoordinates {
	classes := []string{"OUTCOME", "DRIVER", "GUARDRAIL"}
	result := make([]ReaderObservationClassCoordinates, 0, len(classes))
	for _, class := range classes {
		coordinates := ReaderObservationClassCoordinates{Class: class}
		for _, indicator := range indicators {
			if indicator.Class != class {
				continue
			}
			coordinates.Total++
			if indicator.Satisfied {
				coordinates.Satisfied++
			}
		}
		result = append(result, coordinates)
	}
	return result
}

func readerObservationProofs(indicators []ReaderObservationIndicator) []ReaderObservationProofCoordinates {
	proofChoices := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	result := make([]ReaderObservationProofCoordinates, 0, len(proofChoices))
	for _, proofChoice := range proofChoices {
		coordinates := ReaderObservationProofCoordinates{ProofChoice: proofChoice}
		for _, indicator := range indicators {
			if indicator.ProofChoice != proofChoice {
				continue
			}
			coordinates.Total++
			if indicator.Satisfied {
				coordinates.Satisfied++
			}
		}
		result = append(result, coordinates)
	}
	return result
}
