package semantic

import (
	"testing"
)

func TestPROVConformanceCoversKindsAndCoreRelations(t *testing.T) {
	ns := Namespace("bootstrap")
	source := mustEntity(t, MustIdentity("bootstrap://entity/source"), ns, "Source")
	output := mustEntity(t, MustIdentity("bootstrap://entity/output"), ns, "Output")
	compile := mustActivity(t, MustIdentity("bootstrap://activity/compile"), ns, "Compile")
	verify := mustActivity(t, MustIdentity("bootstrap://activity/verify"), ns, "Verify")
	ci := mustAgent(t, GoVerifierID, ns, "Go verifier")
	graph := NewGraph()
	for _, node := range []Node{source, output, compile, verify, ci} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	facts := []Fact{
		NewUsedFact(compile.ID, source.ID),
		NewWasGeneratedByFact(output.ID, compile.ID),
		NewWasDerivedFromFact(output.ID, source.ID),
		NewWasAssociatedWithFact(verify.ID, ci.ID),
	}
	for _, fact := range facts {
		if err := graph.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := len(graph.Facts()); got != len(facts) {
		t.Fatalf("deterministic PROV fact count = %d, want %d", got, len(facts))
	}
}
