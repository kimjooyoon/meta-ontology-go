package bidir

import (
	"testing"
)

func assertEvidenceContract(t *testing.T, evidence BXEvidence, kind ConflictKind) {
	t.Helper()
	if !evidence.GetPutPassed || !evidence.PutGetPassed || !evidence.SemanticEquivalent {
		t.Fatalf("round-trip evidence failed: %#v", evidence)
	}
	if evidence.PartialConflict.Kind != kind || evidence.PartialConflict.Count != 1 {
		t.Fatalf("unexpected conflict evidence: %#v", evidence.PartialConflict)
	}
	if !evidence.PartialConflict.Transactional {
		t.Fatal("conflicting reconciliation was not transactional")
	}
	if evidence.PartialConflict.RemovedCreated || evidence.PartialConflict.CandidatePromoted {
		t.Fatalf("partial observation promoted or removed state: %#v", evidence.PartialConflict)
	}
	if evidence.Delta.CandidatePromoted || !evidence.PartialDelta.PartialObservation {
		t.Fatalf("candidate or partial delta contract was not recorded: %#v", evidence)
	}
	if kind == "" || len(evidence.Delta.Candidates) == 0 && evidence.Delta.EvidenceSpans.IDCount == 0 {
		t.Fatal("canonical delta omitted candidate/evidence records")
	}
	if evidence.RejectedTransaction.Deferred || !evidence.RejectedTransaction.NoWrite || !evidence.PartialConflict.NoWriteObserved {
		t.Fatal("rejected delta transaction was not observer-proven as no-write")
	}
}
func candidateFixtureDelta() FactDelta {
	fact := NewSourcedFact(CandidateFact, "billing://activity/pay-order", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "adapter.go", Start: 1, End: 2})
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = EntityKind
	return FactDelta{Added: FactSet{fact}}
}
func unknownPredicateDelta() FactDelta {
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", Predicate("gooo:unknown"), "billing://entity/order", SourceSpan{File: "adapter.go", Start: 3, End: 4})
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = EntityKind
	return FactDelta{Added: FactSet{fact}}
}
func kindMismatchDelta() FactDelta {
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateUsed, "billing://entity/payment", SourceSpan{File: "adapter.go", Start: 5, End: 6})
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	return FactDelta{Added: FactSet{fact}}
}
