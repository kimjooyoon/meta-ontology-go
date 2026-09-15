package bidir

import (
	"testing"
)

type matrixFixture struct {
	name     string
	accepted FactDelta
	partial  FactDelta
}

func (f matrixFixture) Name() string             { return f.name }
func (f matrixFixture) Document() Document       { return billingDocument() }
func (f matrixFixture) AcceptedDelta() FactDelta { return f.accepted }
func (f matrixFixture) PartialDelta() FactDelta  { return f.partial }
func (f matrixFixture) BaseEvidence() BXBaseEvidenceInput {
	return fixtureBaseEvidence(f.Document())
}
func (f matrixFixture) ObserveAcceptedWrite(before, after Document) BXWriteObservation {
	return fixtureWriteObservation(before, after)
}
func (f matrixFixture) RejectedWriteObserver(document Document) (BXRejectedWriteObserver, error) {
	return NewBXMemoryRejectedWriteObserver(document), nil
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
