package bidir

import (
	"reflect"
	"strings"
	"testing"
)

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
