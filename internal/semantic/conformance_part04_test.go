package semantic

import (
	"testing"
)

func conformanceGraph(t *testing.T, reverse bool) Graph {
	t.Helper()
	ns := Namespace("bootstrap")
	nodes := []Node{
		mustEntity(t, MustIdentity("bootstrap://entity/source"), ns, "Renamed source"),
		mustEntity(t, MustIdentity("bootstrap://entity/output"), ns, "Renamed output"),
		mustActivity(t, MustIdentity("bootstrap://activity/compile"), ns, "Compiler run"),
		mustActivity(t, MustIdentity("bootstrap://activity/verify"), ns, "Verifier run"),
		mustAgent(t, GoVerifierID, ns, "Protected verifier"),
	}
	facts := []Fact{
		NewUsedFact(nodes[2].ID, nodes[0].ID),
		NewWasGeneratedByFact(nodes[1].ID, nodes[2].ID),
		NewWasDerivedFromFact(nodes[1].ID, nodes[0].ID),
		NewWasAssociatedWithFact(nodes[3].ID, nodes[4].ID),
	}
	if reverse {
		reverseNodes(nodes)
		reverseFacts(facts)
	}
	graph := NewGraph()
	for _, node := range nodes {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		if err := graph.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	return graph
}
func reverseNodes(nodes []Node) {
	for left, right := 0, len(nodes)-1; left < right; left, right = left+1, right-1 {
		nodes[left], nodes[right] = nodes[right], nodes[left]
	}
}
func reverseFacts(facts []Fact) {
	for left, right := 0, len(facts)-1; left < right; left, right = left+1, right-1 {
		facts[left], facts[right] = facts[right], facts[left]
	}
}
