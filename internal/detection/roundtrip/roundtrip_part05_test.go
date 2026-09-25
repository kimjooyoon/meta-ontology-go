package roundtrip

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestVerifyDoesNotMutateInputs(t *testing.T) {
	fixture := MinimalFixture()
	beforeGo := append([]byte(nil), fixture.BeforeGo...)
	afterGo := append([]byte(nil), fixture.AfterGo...)
	beforeHash := fixture.IR.StableHash()
	allowed := append([]semantic.ID(nil), fixture.AllowedIDs...)
	if report := Verify(fixture); !report.OK() {
		t.Fatal(report.Error())
	}
	if !bytes.Equal(fixture.BeforeGo, beforeGo) || !bytes.Equal(fixture.AfterGo, afterGo) {
		t.Fatal("verification mutated source bytes")
	}
	if fixture.IR.StableHash() != beforeHash || len(fixture.AllowedIDs) != len(allowed) {
		t.Fatal("verification mutated semantic input")
	}
}
func rebuildIR(t *testing.T, base semantic.IR, transformNode func(semantic.Node) semantic.Node, transformFact func(semantic.Fact) (semantic.Fact, bool)) semantic.IR {
	t.Helper()
	result := semantic.NewIR(base.Package, base.Namespace)
	for _, node := range base.Graph.Nodes() {
		if transformNode != nil {
			node = transformNode(node)
		}
		if err := result.AddNode(node); err != nil {
			t.Fatalf("rebuild node %s: %v", node.ID, err)
		}
	}
	for _, fact := range base.Graph.DeterministicFacts() {
		keep := true
		if transformFact != nil {
			fact, keep = transformFact(fact)
		}
		if keep {
			if err := result.AddFact(fact); err != nil {
				t.Fatalf("rebuild fact %s: %v", fact.Key(), err)
			}
		}
	}
	return result
}
func addNodesAndFacts(t *testing.T, base semantic.IR, nodes []semantic.Node, facts []semantic.Fact) semantic.IR {
	t.Helper()
	result := rebuildIR(t, base, nil, nil)
	for _, node := range nodes {
		if err := result.AddNode(node); err != nil {
			t.Fatalf("add node %s: %v", node.ID, err)
		}
	}
	for _, fact := range facts {
		if err := result.AddFact(fact); err != nil {
			t.Fatalf("add fact %s: %v", fact.Key(), err)
		}
	}
	return result
}
func mustFixtureNode(t *testing.T, kind semantic.Kind, id, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewNodeFromStrings(kind, id, "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
