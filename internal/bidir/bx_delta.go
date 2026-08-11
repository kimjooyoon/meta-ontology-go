package bidir

import "fmt"

func makeDeltaEvidence(delta FactDelta, locality Locality, partial bool, base, after Model) (BXDeltaEvidence, error) {
	evidence := makeDeltaEvidenceUnchecked(delta, locality, partial, base, after)
	if evidence.CanonicalJSON == "" || evidence.LocalityCanonicalJSON == "" {
		return BXDeltaEvidence{}, fmt.Errorf("canonical delta or locality JSON is empty")
	}
	return evidence, nil
}

func makeDeltaEvidenceUnchecked(delta FactDelta, locality Locality, partial bool, base, after Model) BXDeltaEvidence {
	facts := append(append(FactSet{}, delta.Added...), delta.Removed...)
	evidence := BXDeltaEvidence{
		SequenceHash:        factSequenceHash(delta),
		OrderHash:           factOrderHash(delta),
		Locality:            locality,
		LocalityClosureHash: localityDigest(locality),
		EvidenceSpans:       evidenceSpans(facts),
		PartialObservation:  partial,
		RemovedCreated:      removedCreated(base, after, delta),
		CandidatePromoted:   candidatePromoted(base, delta, after),
	}
	evidence.CanonicalJSON = deltaJSON(delta, evidence)
	evidence.LocalityCanonicalJSON = localityJSON(locality, evidence.LocalityClosureHash)
	return evidence
}

func deltaJSON(delta FactDelta, evidence BXDeltaEvidence) string {
	value := struct {
		SequenceHash string   `json:"sequence_hash"`
		OrderHash    string   `json:"order_hash"`
		Added        []string `json:"added"`
		Removed      []string `json:"removed"`
		Partial      bool     `json:"partial_observation"`
	}{
		SequenceHash: evidence.SequenceHash,
		OrderHash:    evidence.OrderHash,
		Added:        factCanonicalValues(delta.Added),
		Removed:      factCanonicalValues(delta.Removed),
		Partial:      evidence.PartialObservation,
	}
	result, _ := canonicalJSON(value)
	return result
}

func localityJSON(locality Locality, closureHash string) string {
	value := struct {
		Touched  []ID   `json:"touched"`
		Affected []ID   `json:"affected"`
		Closure  string `json:"closure_hash"`
	}{Touched: append([]ID{}, locality.Touched...), Affected: append([]ID{}, locality.Affected...), Closure: closureHash}
	result, _ := canonicalJSON(value)
	return result
}

func factCanonicalValues(facts FactSet) []string {
	values := make([]string, len(facts))
	for index, fact := range facts {
		values[index] = factCanonical(fact)
	}
	return values
}
