package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestWrongEndpointAndReversedEdgeFailClosed(t *testing.T) {
	cases := []struct {
		name   string
		mutate func([]semantic.InferenceEdge)
	}{
		{
			name: "wrong endpoint",
			mutate: func(edges []semantic.InferenceEdge) {
				edges[1].ObjectID = fixtureID("node/unregistered-endpoint")
			},
		},
		{
			name: "reversed edge",
			mutate: func(edges []semantic.InferenceEdge) {
				edges[2].SubjectID, edges[2].ObjectID = edges[2].ObjectID, edges[2].SubjectID
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			edges := append([]semantic.InferenceEdge(nil), completeInferenceFixture().edges...)
			test.mutate(edges)
			if err := (semantic.InferencePathV1{
				Version: semantic.InferencePathSchemaVersion, Edges: edges,
				Evidence: completeInferenceFixture().evidence,
			}).Validate(); err != nil {
				t.Fatalf("record bindings rejected before endpoint check: %v", err)
			}
			if _, err := semantic.NewInferencePathChain(edges...); err == nil {
				t.Fatal("malformed endpoint produced a successful chain")
			}
		})
	}
}
func TestTwoPathsWhereOneIsIncompleteAreNotBothComplete(t *testing.T) {
	fixture := completeInferenceFixture()
	complete, err := semantic.NewInferencePathChain(fixture.edges...)
	if err != nil {
		t.Fatalf("complete path rejected: %v", err)
	}
	incompleteEdges := append([]semantic.InferenceEdge(nil), fixture.edges[:1]...)
	incompleteEdges = append(incompleteEdges, fixture.edges[2:]...)
	if _, err := semantic.NewInferencePathChain(incompleteEdges...); err == nil {
		t.Fatal("incomplete alternative produced a successful chain")
	}
	assertExactRecordSequence(t, complete.Edges, fixture.edges)
}
