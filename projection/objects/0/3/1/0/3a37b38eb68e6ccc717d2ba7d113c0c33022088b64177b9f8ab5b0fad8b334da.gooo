package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func newResolutionGraph(t *testing.T, facts []Fact) *Graph {
	t.Helper()
	graph := New()
	nodes := []Node{
		{ID: id("urn:resolution:entity:business"), Kind: EntityNodeKind},
		{ID: id("urn:resolution:activity:det"), Kind: ActivityNodeKind},
		{ID: id("urn:resolution:activity:candidate"), Kind: ActivityNodeKind},
		{ID: id("urn:resolution:entity:generated-det"), Kind: EntityNodeKind},
		{ID: id("urn:resolution:entity:generated-candidate"), Kind: EntityNodeKind},
	}
	for _, node := range nodes {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		assertAdd(t, graph, fact)
	}
	return graph
}
func resolutionLabel(metadata EnvelopeMetadata) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == "resolution_view" {
			return label
		}
	}
	return AuthorityLabel{}
}
func mustSemanticEntity(t *testing.T, raw, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewEntity(semantic.MustIdentity(raw), "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
func mustSemanticActivity(t *testing.T, raw, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewActivity(semantic.MustIdentity(raw), "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
