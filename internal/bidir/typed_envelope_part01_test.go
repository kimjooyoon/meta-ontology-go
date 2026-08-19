package bidir

import (
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
