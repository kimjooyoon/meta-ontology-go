package externalcapabilityexecution

func makeProofs(indicators []Indicator) []Proof {
	proofs := make([]Proof, 0, 3)
	for _, mode := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		proof := Proof{Mode: mode, Status: StatusSatisfied, Resolution: ResolutionExact}
		for _, metric := range indicators {
			if metric.ProofChoice != mode {
				continue
			}
			proof.Total++
			if metric.Status == StatusSatisfied {
				proof.Completed++
			}
			if metric.Status == StatusUnsatisfied && proof.Status != StatusUnknown {
				proof.Status, proof.Resolution = StatusUnsatisfied, ResolutionInvariant
			}
			if metric.Status == StatusUnknown {
				proof.Status, proof.Resolution = StatusUnknown, ResolutionUnknown
			}
		}
		proofs = append(proofs, proof)
	}
	return proofs
}
