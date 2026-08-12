package bidir

import "testing"

type matrixFixture struct {
	name     string
	accepted FactDelta
	partial  FactDelta
}

func (f matrixFixture) Name() string { return f.name }

func (f matrixFixture) Document() Document { return billingDocument() }

func (f matrixFixture) AcceptedDelta() FactDelta { return f.accepted }

func (f matrixFixture) PartialDelta() FactDelta { return f.partial }

func (f matrixFixture) BaseEvidence() BXBaseEvidenceInput {
	return fixtureBaseEvidence(f.Document())
}

func (f matrixFixture) ObserveAcceptedWrite(before, after Document) BXWriteObservation {
	return fixtureWriteObservation(before, after)
}

func TestBXEvidenceMatrixCoversConflictClasses(t *testing.T) {
	candidate := candidateFixtureDelta()
	cases := []struct {
		name    string
		fixture ReconciliationFixture
		kind    ConflictKind
	}{
		{name: "missing-source", fixture: billingBXFixture{}, kind: ConflictMissingSource},
		{name: "unknown-predicate", fixture: matrixFixture{
			name:     "unknown-predicate",
			accepted: candidate,
			partial:  unknownPredicateDelta(),
		}, kind: ConflictUnknownPredicate},
		{name: "kind-mismatch", fixture: matrixFixture{
			name:     "kind-mismatch",
			accepted: candidate,
			partial:  kindMismatchDelta(),
		}, kind: ConflictKindMismatch},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			evidence, err := MeasureBXFixture(testCase.fixture)
			if err != nil {
				t.Fatal(err)
			}
			assertEvidenceContract(t, evidence, testCase.kind)
		})
	}
}

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
	if !evidence.RejectedTransaction.Deferred || evidence.PartialConflict.NoWriteObserved {
		t.Fatal("rejected delta transaction was not explicitly deferred")
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
