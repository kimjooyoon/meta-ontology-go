package authorization

func makeProofs(indicators []Indicator) []Proof {
	proofs := make([]Proof, 0, 3)
	for _, choice := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		proof := Proof{Mode: choice, Status: StatusSatisfied, Resolution: ResolutionExact}
		for _, indicator := range indicators {
			if indicator.ProofChoice != choice {
				continue
			}
			proof.Total++
			if indicator.Status == StatusSatisfied {
				proof.Completed++
			}
			if indicator.Status == StatusUnsatisfied && proof.Status != StatusUnknown {
				proof.Status = StatusUnsatisfied
			}
			if indicator.Status == StatusUnknown {
				proof.Status, proof.Resolution = StatusUnknown, ResolutionUnknown
			}
		}
		proofs = append(proofs, proof)
	}
	return proofs
}
