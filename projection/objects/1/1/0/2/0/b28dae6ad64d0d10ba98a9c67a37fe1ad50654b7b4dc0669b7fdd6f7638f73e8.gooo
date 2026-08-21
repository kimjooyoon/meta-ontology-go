package query

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

func TestEnvelopeReplayAndPermutationAreCanonical(t *testing.T) {
	facts := []Fact{
		NewFact(id("urn:gooo:activity:pay"), Used, id("urn:gooo:entity:order")),
		NewCandidateFact(id("urn:gooo:entity:order"), WasDerivedFrom, id("urn:gooo:entity:archive"), "ambiguous"),
		NewFact(id("urn:gooo:entity:payment"), WasGeneratedBy, id("urn:gooo:activity:pay")),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for _, fact := range slices.Backward(facts) {
		assertAdd(t, second, fact)
	}
	request := traversalEnvelope(id("urn:gooo:activity:pay"), LayerAll, 2, 10)
	request.Relation = PROVUsed
	beforeCanonical, beforeNodes := first.Canonical(), first.Nodes()
	want, err := first.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		got, err := second.Execute(request)
		if err != nil {
			t.Fatal(err)
		}
		assertSameEnvelope(t, got, want)
	}
	canonical, err := want.CanonicalJSON()
	if err != nil || !json.Valid(canonical) {
		t.Fatalf("invalid canonical response JSON: %s %v", canonical, err)
	}
	if want.Hash == "" || want.Hash != want.CanonicalDigestValue() {
		t.Fatalf("response hash is not a receipt: %q", want.Hash)
	}
	canonicalRequest, err := request.CanonicalDigest()
	if err != nil || canonicalRequest != want.RequestHash {
		t.Fatalf("request hash is not canonical: %q/%q", canonicalRequest, want.RequestHash)
	}
	if first.Canonical() != beforeCanonical || !reflect.DeepEqual(first.Nodes(), beforeNodes) {
		t.Fatal("successful envelope query mutated the graph")
	}
	if want.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("missing provenance was not deferred: %#v", want.Metadata)
	}
	if label := authorityLabel(want.Metadata, "provenance"); label.Status != StatusDeferred {
		t.Fatalf("provenance authority label = %#v", label)
	}
	if label := authorityLabel(want.Metadata, "query_graph"); label.Authority != "derived" {
		t.Fatalf("query graph authority label = %#v", label)
	}
	var wire map[string]any
	if err := json.Unmarshal(mustMarshal(t, want), &wire); err != nil {
		t.Fatal(err)
	}
	if wire["schema"] != QueryEnvelopeSchema || wire["canonical_hash"] != want.Hash {
		t.Fatalf("wire envelope identity = %#v", wire)
	}
}
