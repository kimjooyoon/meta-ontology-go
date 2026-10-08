package bidir

import (
	"os"
	"testing"
)

func TestQuadraticSearchAlternativeSurvivesSemanticRoundTrip(t *testing.T) {
	source, err := os.ReadFile("../../examples/quadratic-feedback/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	document := assemblyDocument(t, source)
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Get(written)
	if err != nil || !SemanticEquivalent(model, again) {
		t.Fatal("quadratic alternative changed in semantic round-trip", err)
	}
}
