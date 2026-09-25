package artifactemit

func symbolicValueCoordinates(indicators []SymbolicValueContractIndicator) SymbolicValueContractCoordinates {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	total := len(indicators)
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10000 / total
	}
	return SymbolicValueContractCoordinates{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func symbolicValueClasses(indicators []SymbolicValueContractIndicator) []SymbolicValueContractClass {
	classes := []string{"OUTCOME", "DRIVER", "GUARDRAIL"}
	result := make([]SymbolicValueContractClass, 0, len(classes))
	for _, class := range classes {
		total := 0
		satisfied := 0
		for _, indicator := range indicators {
			if indicator.Class != class {
				continue
			}
			total++
			if indicator.Satisfied {
				satisfied++
			}
		}
		result = append(result, SymbolicValueContractClass{Class: class, Satisfied: satisfied, Total: total})
	}
	return result
}
