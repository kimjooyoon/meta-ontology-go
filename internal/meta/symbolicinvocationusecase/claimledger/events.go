package claimledger

func addEvent(report *Report, spec ClaimSpec, kind, status, reason, evidenceID string) {
	report.Events = append(report.Events, Event{
		Sequence: len(report.Events) + 1, Type: kind, ClaimID: spec.ID,
		EvidenceID: evidenceID, Status: status, Coordinate: spec.Coordinate, Reason: reason,
	})
}

func countProofRoute(counts *ProofRouteCounts, route string) {
	switch route {
	case "FOUNDATION":
		counts.Foundation++
	case "COHERENCE":
		counts.Coherence++
	case "REGRESSION":
		counts.Regression++
	}
}
