package toolchainrelease

type CaseResult struct {
	ID             string `json:"id"`
	TargetID       string `json:"target_id,omitempty"`
	Kind           string `json:"kind"`
	Expected       string `json:"expected"`
	Observed       string `json:"observed"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type caseObservation struct {
	Observed string
	Digest   string
	Ready    bool
}

func evaluateCases(corpus Corpus, observations map[string]caseObservation) ([]CaseResult, int) {
	results := make([]CaseResult, 0, len(corpus.Cases))
	satisfied := 0
	for _, expected := range corpus.Cases {
		observation := observations[expected.ID]
		status := CaseNotSatisfied
		if observation.Ready && observation.Observed == expected.Expected {
			status = CaseSatisfied
			satisfied++
		}
		results = append(results, CaseResult{
			ID: expected.ID, TargetID: expected.TargetID, Kind: expected.Kind,
			Expected: expected.Expected, Observed: observation.Observed,
			Status: status, EvidenceDigest: observation.Digest,
		})
	}
	return results, satisfied
}
