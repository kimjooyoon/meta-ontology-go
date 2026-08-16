package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestTypedEnvelopeEmitsBoundedLawBundle(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	result, err := AdaptBXTypedEnvelope(projection, observation)
	if err != nil {
		t.Fatal(err)
	}
	if result.FeatureGreen || !result.GetPutPassed || !result.PutGetPassed || !result.SemanticEquivalent {
		t.Fatalf("typed laws overclaimed or failed: %#v", result)
	}
	if !result.ThreeWay.Succeeded() || !SemanticEquivalent(result.ThreeWay.Model, projection.Left) {
		t.Fatal("three-way result was not emitted as the accepted semantic model")
	}
	if result.Delta.CanonicalJSON == "" || result.PartialDelta.LocalityCanonicalJSON == "" || result.Candidates == nil || result.PortSequence == nil || result.RelationSequence == nil {
		t.Fatalf("typed bundle omitted delta/locality/order/candidate evidence: %#v", result)
	}
	if result.ObserverKind != "memory-source" || !result.NoWriteObserved || result.PartialDelta.RemovedCreated || result.PartialDelta.CandidatePromoted {
		t.Fatalf("typed bundle overclaimed rejected observation: %#v", result)
	}
	if !strings.Contains(result.CanonicalJSON, `"feature_green":false`) || result.Hash == "" {
		t.Fatalf("typed canonical evidence is incomplete: %s", result.CanonicalJSON)
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestTypedEnvelopeRejectsMissingObserverReceipt(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	observation.Rejected = nil
	if _, err := AdaptBXTypedEnvelope(projection, observation); err == nil || !strings.Contains(err.Error(), "observer receipt is absent") {
		t.Fatalf("missing observer receipt was accepted: %v", err)
	}
}

func TestTypedEnvelopeRejectsStaleProjection(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	projection.Base.Relations = append(projection.Base.Relations, Relation{
		Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment",
	})
	projection, err := SealBXTypedProjectionEnvelope(projection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AdaptBXTypedEnvelope(projection, observation); err == nil || !strings.Contains(err.Error(), "base model is stale") {
		t.Fatalf("stale projection was accepted: %v", err)
	}
}

func TestTypedEnvelopeRejectsTamperedProjection(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	projection.Fixture = "relabelled-fixture"
	if _, err := AdaptBXTypedEnvelope(projection, observation); err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("tampered projection was accepted: %v", err)
	}
}

func TestTypedEnvelopeRejectsRelabeledObserverReceipt(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	receipt := *observation.Rejected
	receipt.ObserverKind = "filesystem/inode"
	observation.Rejected = &receipt
	if _, err := AdaptBXTypedEnvelope(projection, observation); err == nil || !strings.Contains(err.Error(), "receipt digest mismatch") {
		t.Fatalf("relabeled observer receipt was accepted: %v", err)
	}
}

func TestTypedEnvelopeRejectsTamperedResultCanonical(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	result, err := AdaptBXTypedEnvelope(projection, observation)
	if err != nil {
		t.Fatal(err)
	}
	result.CanonicalJSON = strings.Replace(result.CanonicalJSON, `"feature_green":false`, `"feature_green":true`, 1)
	result.Hash = digest(result.CanonicalJSON)
	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), "canonical evidence is stale") {
		t.Fatalf("tampered result canonical evidence was accepted: %v", err)
	}
}

func TestTypedEnvelopePartialObservationIsNoDeleteAndNoWrite(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	before := projection.Base.Clone()
	result, err := AdaptBXTypedEnvelope(projection, observation)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Evidence.PartialConflict.Transactional || !result.Evidence.PartialConflict.NoWriteObserved {
		t.Fatalf("partial observation was not transactional/no-write: %#v", result.Evidence.PartialConflict)
	}
	if len(result.PartialDelta.Removed) != 0 || result.PartialDelta.RemovedCreated || result.PartialDelta.CandidatePromoted {
		t.Fatalf("partial observation created a deletion or promotion: %#v", result.PartialDelta)
	}
	if !SemanticEquivalent(before, projection.Base) || !result.NoWriteObserved {
		t.Fatal("typed adapter mutated input or lost no-write evidence")
	}
}

func TestTypedEnvelopeReplayIsStableAndDetached(t *testing.T) {
	projection, observation := typedEnvelopeFixture(t)
	projectionBefore, observationBefore := projection, observation
	first, err := AdaptBXTypedEnvelope(projection, observation)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AdaptBXTypedEnvelope(projection, observation)
	if err != nil {
		t.Fatal(err)
	}
	if first.Hash != second.Hash || first.CanonicalJSON != second.CanonicalJSON || !reflect.DeepEqual(first.Evidence, second.Evidence) {
		t.Fatal("typed envelope replay was not deterministic")
	}
	if !reflect.DeepEqual(projection, projectionBefore) || !reflect.DeepEqual(observation, observationBefore) {
		t.Fatal("typed adapter mutated an input envelope")
	}
	first.Candidates = append(first.Candidates, "tampered")
	if reflect.DeepEqual(first, second) {
		t.Fatal("typed result slices were not detached")
	}
}

func typedEnvelopeFixture(t *testing.T) (BXTypedProjectionEnvelope, BXTypedObservationEnvelope) {
	t.Helper()
	fixture := billingBXFixture{}
	document := fixture.Document()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := Reconcile(base, fixture.AcceptedDelta())
	if err != nil {
		t.Fatal(err)
	}
	updated, err := Put(document, accepted.Model)
	if err != nil {
		t.Fatal(err)
	}
	observer := NewBXMemoryRejectedWriteObserver(document)
	receipt, err := CaptureBXObserverReceipt(observer, func() error {
		_, reconcileErr := Reconcile(base, fixture.PartialDelta())
		return reconcileErr
	})
	if err != nil {
		t.Fatal(err)
	}
	projection := BXTypedProjectionEnvelope{
		SchemaVersion: BXTypedEnvelopeSchemaVersion, Fixture: "typed-billing", Document: document,
		Base: base, Left: accepted.Model, Right: base.Clone(), AcceptedDelta: fixture.AcceptedDelta(),
		PartialDelta: fixture.PartialDelta(), BaseEvidence: fixture.BaseEvidence(),
	}
	projection, err = SealBXTypedProjectionEnvelope(projection)
	if err != nil {
		t.Fatal(err)
	}
	observation := NewBXTypedObservationEnvelope(fixtureWriteObservation(document, updated), receipt)
	return projection, observation
}
