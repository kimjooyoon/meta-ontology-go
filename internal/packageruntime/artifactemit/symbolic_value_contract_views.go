package artifactemit

import "slices"

func symbolicValueView(indicators []SymbolicValueContractIndicator, audience, resolution string) SymbolicValueContractView {
	total := 0
	satisfied := 0
	for _, indicator := range indicators {
		if !slices.Contains(indicator.Audiences, audience) {
			continue
		}
		total++
		if indicator.Satisfied {
			satisfied++
		}
	}
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10000 / total
	}
	return SymbolicValueContractView{
		Audience: audience, Resolution: resolution, Satisfied: satisfied, Total: total, BasisPoints: basisPoints,
	}
}

func symbolicValueProofs(indicators []SymbolicValueContractIndicator) []SymbolicValueContractProof {
	proofs := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	result := make([]SymbolicValueContractProof, 0, len(proofs))
	for _, proof := range proofs {
		total := 0
		satisfied := 0
		for _, indicator := range indicators {
			if indicator.ProofChoice != proof {
				continue
			}
			total++
			if indicator.Satisfied {
				satisfied++
			}
		}
		result = append(result, SymbolicValueContractProof{ProofChoice: proof, Satisfied: satisfied, Total: total})
	}
	return result
}
