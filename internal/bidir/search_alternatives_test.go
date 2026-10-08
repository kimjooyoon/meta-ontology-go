package bidir

import (
	"os"
	"strings"
	"testing"
)

func TestSearchAlternativesSurviveSemanticRoundTrip(t *testing.T) {
	source, err := os.ReadFile("../../examples/search-feedback/source.gooo.fixture")
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
		t.Fatal("search alternatives changed under Get-Put-Get", err)
	}
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := LowerDocument(assemblyDocument(t, []byte(strings.Replace(string(source), `max_candidates "16"`, `max_candidates "8"`, 1))))
	if err != nil || changed.StableHash() == core.StableHash() {
		t.Fatal("alternative cap did not contribute to source meaning", err)
	}
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Add" {
			continue
		}
		if node.Assembly == nil || len(node.Assembly.SearchAlternatives) != 2 {
			t.Fatal("source alternatives were lost during lowering")
		}
		node.Assembly.SearchAlternatives[0].ID = "changed"
		fresh, _ := core.Graph.Node(node.ID)
		if fresh.Assembly.SearchAlternatives[0].ID == "changed" {
			t.Fatal("semantic lookup shares mutable alternative storage")
		}
		return
	}
	t.Fatal("activity identity missing")
}
