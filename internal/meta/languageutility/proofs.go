package languageutility

func buildProofs(cells []CellResult, evidenceDigest string) []Proof {
	choices := []string{"foundation", "coherence", "regression"}
	result := make([]Proof, 0, len(choices))
	for _, choice := range choices {
		proof := Proof{Choice: choice, EvidenceDigest: evidenceDigest}
		unknown, refuted := 0, 0
		for _, cell := range cells {
			if cell.ProofChoice != choice {
				continue
			}
			proof.Total++
			if cell.State == StateClosed {
				proof.Closed++
			}
			if cell.State == StateUnknown {
				unknown++
			}
			if cell.State == StateRefuted {
				refuted++
			}
		}
		proof.Status = "GAP"
		if unknown > 0 {
			proof.Status = "UNKNOWN"
		} else if refuted > 0 {
			proof.Status = "REFUTED"
		} else if proof.Closed == proof.Total {
			proof.Status = "SATISFIED"
		}
		result = append(result, proof)
	}
	return result
}
