package semantic

import (
	"errors"
	"testing"
)

func TestPROVAttributionIsExplicitlyUnsupported(t *testing.T) {
	relation := Relation("wasAttributedTo")
	if relation.Valid() {
		t.Fatal("unsupported attribution relation became valid implicitly")
	}
	graph := NewGraph()
	fact := NewFact(
		MustIdentity("prov-guard://entity/output"), relation,
		MustIdentity("prov-guard://agent/owner"),
	)
	if err := graph.AddFact(fact); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("unsupported attribution error = %v, want ErrUnknownRelation", err)
	}
	if len(graph.AllFacts()) != 0 {
		t.Fatal("rejected attribution mutated graph")
	}
}
func candidateHashIR(t *testing.T, includeCandidate, reverse bool) IR {
	t.Helper()
	ns := Namespace("candidate-hash")
	ir := NewIR("candidate-hash", ns)
	nodes := []Node{
		mustEntity(t, MustIdentity("candidate-hash://entity/source"), ns, "Source"),
		mustEntity(t, MustIdentity("candidate-hash://entity/output"), ns, "Output"),
		mustActivity(t, MustIdentity("candidate-hash://activity/compile"), ns, "Compile"),
	}
	facts := []Fact{NewUsedFact(nodes[2].ID, nodes[0].ID)}
	if includeCandidate {
		facts = append(facts, candidateHashFact())
	}
	if reverse {
		reverseNodes(nodes)
		reverseFacts(facts)
	}
	for _, node := range nodes {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		if includeCandidate && fact.Status == FactCandidate {
			if err := ir.AddCandidate(fact); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}
func candidateHashFact() Fact {
	return NewCandidateFact(
		MustIdentity("candidate-hash://activity/compile"), Used,
		MustIdentity("candidate-hash://entity/output"),
		"observed output dependency",
	)
}
