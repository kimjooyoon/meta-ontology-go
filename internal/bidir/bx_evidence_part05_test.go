package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestBXRejectedObserverOwnsNoWriteSnapshots(t *testing.T) {
	document := billingDocument()
	observer := NewBXMemoryRejectedWriteObserver(document)
	called := false
	observation, err := observer.ObserveRejected(func() error {
		called = true
		return nil
	})
	if err != nil || !called || observer.Kind() != "memory-source" {
		t.Fatalf("observer did not run its operation: err=%v called=%t kind=%q", err, called, observer.Kind())
	}
	if !observation.Observed || !reflect.DeepEqual(observation.Before, observation.After) {
		t.Fatalf("observer did not prove no-write: %#v", observation)
	}
	observation.Before.Bytes[0] ^= 1
	second, err := observer.ObserveRejected(func() error { return nil })
	if err != nil || !reflect.DeepEqual(second.Before, second.After) {
		t.Fatalf("observer snapshots were not isolated: err=%v observation=%#v", err, second)
	}
}
func TestBXDeltaEvidenceRetainsRemovedCandidates(t *testing.T) {
	candidate := NewSourcedFact(CandidateFact, "billing://activity/pay-order", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "removed.go", Start: 1, End: 2})
	evidence := makeDeltaEvidenceUnchecked(FactDelta{Removed: FactSet{candidate}}, Locality{}, true, Model{}, Model{})
	if len(evidence.Candidates) != 1 || !strings.Contains(evidence.CanonicalJSON, "\"candidates\"") {
		t.Fatalf("removed candidate was omitted from canonical evidence: %#v", evidence)
	}
}

type incompleteBXFixture struct{}

func (incompleteBXFixture) Name() string             { return "incomplete" }
func (incompleteBXFixture) Document() Document       { return billingDocument() }
func (incompleteBXFixture) AcceptedDelta() FactDelta { return FactDelta{} }
func (incompleteBXFixture) PartialDelta() FactDelta  { return FactDelta{} }

type missingArtifactsFixture struct{ billingBXFixture }

func (missingArtifactsFixture) Name() string { return "missing-artifacts" }
func (missingArtifactsFixture) BaseEvidence() BXBaseEvidenceInput {
	return BXBaseEvidenceInput{DSL: billingDocument()}
}
