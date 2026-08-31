package rollbackfixedpoint

func proofs(coordinates []Coordinate) []Proof {
	return []Proof{
		proof("FOUNDATION", "bind-canonical-cycle-evidence", coordinates[0:3]),
		proof("COHERENCE", "cohere-guard-and-transformation-outcome", coordinates[3:6]),
		proof("REGRESSION", "reject-effects-writes-or-authority", coordinates[6:10]),
	}
}

func proof(choice, operation string, values []Coordinate) Proof {
	passed := true
	for _, value := range values {
		passed = passed && value.Status == "SATISFIED"
	}
	return Proof{Choice: choice, MetaOperation: operation, Passed: passed,
		EvidenceDigest: digestJSON(values)}
}
