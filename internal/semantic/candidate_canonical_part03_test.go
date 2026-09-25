package semantic

import (
	"errors"
	"testing"
)

func TestRejectedFactAndNodeOperationsDoNotMutateGraph(t *testing.T) {
	ns := Namespace("candidate-hash")
	activity := mustActivity(t, MustIdentity("candidate-hash://activity/compile"), ns, "Compile")
	entity := mustEntity(t, MustIdentity("candidate-hash://entity/source"), ns, "Source")
	output := mustEntity(t, MustIdentity("candidate-hash://entity/output"), ns, "Output")
	verify := mustActivity(t, MustIdentity("candidate-hash://activity/verify"), ns, "Verify")
	agent := mustAgent(t, MustIdentity("candidate-hash://agent/verifier"), ns, "Verifier")
	cases := []struct {
		name string
		add  func(*Graph) error
		want error
	}{
		{"used direction", func(g *Graph) error { return g.AddFact(NewUsedFact(entity.ID, activity.ID)) }, ErrInvalidFact},
		{"generated direction", func(g *Graph) error { return g.AddFact(NewWasGeneratedByFact(activity.ID, output.ID)) }, ErrInvalidFact},
		{"derived direction", func(g *Graph) error { return g.AddFact(NewWasDerivedFromFact(output.ID, activity.ID)) }, ErrInvalidFact},
		{"associated direction", func(g *Graph) error { return g.AddFact(NewWasAssociatedWithFact(verify.ID, output.ID)) }, ErrInvalidFact},
		{"kind rekey", func(g *Graph) error {
			return g.AddNode(Node{ID: activity.ID, Kind: Entity, Namespace: ns, Name: "Compile"})
		}, ErrIdentityConflict},
		{"namespace rekey", func(g *Graph) error {
			return g.AddNode(Node{ID: activity.ID, Kind: Activity, Namespace: Namespace("other"), Name: "Compile"})
		}, ErrIdentityConflict},
		{"name collision", func(g *Graph) error {
			return g.AddNode(Node{ID: MustIdentity("candidate-hash://entity/other"), Kind: Entity, Namespace: ns, Name: "Source"})
		}, ErrNameCollision},
		{"alias collision", func(g *Graph) error {
			return g.AddNode(Node{ID: MustIdentity("candidate-hash://entity/alias"), Kind: Entity, Namespace: ns, Name: "Other", Aliases: []string{"Source"}})
		}, ErrNameCollision},
	}
	for _, test := range cases {
		graph := NewGraph()
		for _, node := range []Node{activity, entity, output, verify, agent} {
			if err := graph.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		before := graph.Canonical()
		if err := test.add(&graph); !errors.Is(err, test.want) {
			t.Errorf("%s error = %v, want %v", test.name, err, test.want)
		}
		if graph.Canonical() != before {
			t.Errorf("%s mutated graph after rejection", test.name)
		}
	}
}
func TestNamespaceQualifiedNamesAllowDistinctNamespaces(t *testing.T) {
	firstNS := Namespace("first")
	secondNS := Namespace("second")
	graph := NewGraph()
	for _, node := range []Node{
		mustEntity(t, MustIdentity("namespace://entity/first"), firstNS, "Record"),
		mustEntity(t, MustIdentity("namespace://entity/second"), secondNS, "Record"),
	} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, ok := graph.NodeByName(firstNS, "Record"); !ok {
		t.Fatal("first namespace-qualified name did not resolve")
	}
	if _, ok := graph.NodeByName(secondNS, "Record"); !ok {
		t.Fatal("second namespace-qualified name did not resolve")
	}
}
