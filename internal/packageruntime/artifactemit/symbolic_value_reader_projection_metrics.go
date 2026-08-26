package artifactemit

func symbolicReaderCoordinates(satisfied, total int) SymbolicValueContractCoordinates {
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10000 / total
	}
	return SymbolicValueContractCoordinates{
		Satisfied: satisfied, Total: total, BasisPoints: basisPoints,
	}
}

func symbolicReaderMetricCoordinates(
	indicators []SymbolicValueContractIndicator,
) SymbolicValueContractCoordinates {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	return symbolicReaderCoordinates(satisfied, len(indicators))
}

func symbolicReaderClasses(
	indicators []SymbolicValueContractIndicator,
) []SymbolicValueContractClass {
	classes := []string{"OUTCOME", "DRIVER", "GUARDRAIL"}
	result := make([]SymbolicValueContractClass, 0, len(classes))
	for _, class := range classes {
		satisfied, total := 0, 0
		for _, indicator := range indicators {
			if indicator.Class == class {
				total++
				if indicator.Satisfied {
					satisfied++
				}
			}
		}
		result = append(result, SymbolicValueContractClass{Class: class, Satisfied: satisfied, Total: total})
	}
	return result
}

func symbolicReaderProofs(
	indicators []SymbolicValueContractIndicator,
) []SymbolicValueContractProof {
	choices := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	result := make([]SymbolicValueContractProof, 0, len(choices))
	for _, choice := range choices {
		satisfied, total := 0, 0
		for _, indicator := range indicators {
			if indicator.ProofChoice == choice {
				total++
				if indicator.Satisfied {
					satisfied++
				}
			}
		}
		result = append(result, SymbolicValueContractProof{ProofChoice: choice, Satisfied: satisfied, Total: total})
	}
	return result
}
