package bidir

import (
	"fmt"
)

func makeDeltaEvidence(delta FactDelta, locality Locality, partial bool, base, after Model) (BXDeltaEvidence, error) {
	evidence := makeDeltaEvidenceUnchecked(delta, locality, partial, base, after)
	if evidence.CanonicalJSON == "" || evidence.LocalityCanonicalJSON == "" {
		return BXDeltaEvidence{}, fmt.Errorf("canonical delta or locality JSON is empty")
	}
	return evidence, nil
}
func makeDeltaEvidenceUnchecked(delta FactDelta, locality Locality, partial bool, base, after Model) BXDeltaEvidence {

	facts := append(append(FactSet{}, delta.Added...), delta.Removed...)
	ports, relations := orderedSequences(after)
	portHash, relationHash := sequenceHash(ports), sequenceHash(relations)
	sequenceHash := factSequenceHash(delta)
	closure := LocalityBetween(base, after)
	evidenceSet := evidenceSpans(facts)
	evidence := BXDeltaEvidence{
		SequenceHash:        sequenceHash,
		OrderHash:           deltaOrderHash(sequenceHash, portHash, relationHash),
		Locality:            detachedLocality(locality),
		Added:               factCanonicalValues(delta.Added),
		Removed:             factCanonicalValues(delta.Removed),
		LocalityClosureHash: localityDigest(locality),
		ClosureMembers:      append([]ID{}, locality.Affected...),
		ClosureValid:        sameLocality(locality, closure),
		Candidates:          factCanonicalValues(candidateFacts(facts)),
		PortSequence:        ports,
		RelationSequence:    relations,
		PortOrderHash:       portHash,
		RelationOrderHash:   relationHash,
		EvidenceSpans:       evidenceSet,
		EvidenceHash:        evidenceSet.Hash,
		PartialObservation:  partial,
		RemovedCreated:      removedCreated(base, after, delta),
		CandidatePromoted:   candidatePromoted(base, delta, after),
	}
	evidence.CanonicalJSON = deltaJSON(delta, evidence)
	evidence.LocalityCanonicalJSON = localityJSON(locality, evidence.LocalityClosureHash)
	return evidence
}

// deltaOrderHash binds the observed fact sequence to the source-authoritative
// semantic collection orders. All inputs are fixed-width SHA-256 values, so
// the delimiter is unambiguous and the hash can be recomputed from the
// detached evidence record alone.
func deltaOrderHash(sequenceHash, portHash, relationHash string) string {
	return digest(sequenceHash + "|" + portHash + "|" + relationHash)
}
