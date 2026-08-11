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
