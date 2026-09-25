package proposalcompat

func proofs(coordinates []Coordinate) []Proof {
	return []Proof{
		proof("FOUNDATION", "bind-v2-promotion-source", coordinates[0:2]),
		proof("COHERENCE", "project-v2-source-to-v1-target", coordinates[2:5]),
		proof("REGRESSION", "reject-lossy-or-writing-projection", coordinates[5:6]),
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
