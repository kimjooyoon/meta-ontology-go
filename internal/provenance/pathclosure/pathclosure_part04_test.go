package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestEvaluatorTreatsMissingAndAmbiguousEvidenceAsFullSuiteUnknown(t *testing.T) {
	fixture := completeInferenceFixture()
	first := requirementForPath("first", fixture.edges)
	second := requirementForPath("second", fixture.edges)

	cases := []struct {
		name   string
		mutate func(*semantic.InferencePathV1)
	}{
		{
			name: "orphan evidence",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[0].Evidence[0].ID = fixtureID("evidence/not-present")
			},
		},
		{
			name: "ambiguous evidence",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[0].Evidence = append(path.Edges[0].Evidence, path.Edges[0].Evidence[0])
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := clonePath(fixture.path)
			test.mutate(&path)
			got := pathclosure.Evaluate(path, []pathclosure.Requirement{first, second})
			assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingEvidence, 0, 2)
			if !reflect.DeepEqual(got.Missing, []semantic.ID{first.PathID, second.PathID}) {
				t.Fatalf("full-suite missing paths = %#v", got.Missing)
			}
		})
	}
}
