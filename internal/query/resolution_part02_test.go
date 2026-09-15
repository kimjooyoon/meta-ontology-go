package query

import (
	"reflect"
	"testing"
)

func TestGeneratedResolutionPermutationAndCandidateSeparation(t *testing.T) {
	business := id("urn:resolution:entity:business")
	activityDet := id("urn:resolution:activity:det")
	activityCandidate := id("urn:resolution:activity:candidate")
	generatedDet := id("urn:resolution:entity:generated-det")
	generatedCandidate := id("urn:resolution:entity:generated-candidate")
	facts := []Fact{
		NewFact(activityDet, Used, business),
		NewFact(generatedDet, WasGeneratedBy, activityDet),
		NewCandidateFact(activityCandidate, Used, business, "ambiguous activity"),
		NewCandidateFact(generatedCandidate, WasGeneratedBy, activityCandidate, "ambiguous output"),
	}
	first := newResolutionGraph(t, facts)
	secondFacts := append([]Fact(nil), facts...)
	for left, right := 0, len(secondFacts)-1; left < right; left, right = left+1, right-1 {
		secondFacts[left], secondFacts[right] = secondFacts[right], secondFacts[left]
	}
	second := newResolutionGraph(t, secondFacts)
	request := resolutionRequest(business, LayerAll, 10)
	want, err := first.ResolveGeneratedCode(request)
	if err != nil {
		t.Fatal(err)
	}
	got, err := second.ResolveGeneratedCode(request)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := want.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	gotJSON, err := got.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantJSON, gotJSON) || want.Hash != got.Hash {
		t.Fatalf("permutation changed resolution: %q/%q", want.Hash, got.Hash)
	}
	if len(want.Deterministic) != 1 || len(want.Candidates) != 1 ||
		want.Candidates[0].SourceLayer != FactCandidate.String() {
		t.Fatalf("resolution layers leaked: %#v/%#v", want.Deterministic, want.Candidates)
	}
	candidate, err := first.ResolveGeneratedCode(resolutionRequest(business, LayerCandidate, 10))
	if err != nil || len(candidate.Deterministic) != 0 || len(candidate.Candidates) != 1 {
		t.Fatalf("candidate resolution = %#v %v", candidate, err)
	}
	limited, err := first.ResolveGeneratedCode(resolutionRequest(business, LayerAll, 1))
	if err != nil || len(limited.Deterministic) != 1 || len(limited.Candidates) != 0 {
		t.Fatalf("resolution limit did not prefer deterministic: %#v %v", limited, err)
	}
}
