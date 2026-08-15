package pathclosure_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func assertValidatorError(t *testing.T, path semantic.InferencePathV1, detail string) {
	t.Helper()
	err := path.Validate()
	if err == nil || !errors.Is(err, semantic.ErrInferencePath) || !strings.Contains(err.Error(), detail) {
		t.Fatalf("validator error = %v, want ErrInferencePath containing %q", err, detail)
	}
}

func TestMissingEvidenceAndControlInputsAreRejectedBySemanticValidator(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*semantic.InferencePathV1)
		want   string
	}{
		{
			name: "missing evidence",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[1].Evidence = nil
			},
			want: "at least one evidence reference is required",
		},
		{
			name: "missing projection profile",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[2].Controls.Profile = semantic.ProfileBinding{}
			},
			want: "profile binding is required",
		},
		{
			name: "missing verification policy",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[3].Controls.PolicyDigest = ""
			},
			want: "policy digest is required",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := clonePath(completeInferenceFixture().path)
			test.mutate(&path)
			assertValidatorError(t, path, test.want)
		})
	}
}

func TestDuplicateRecordAndWrongKindFailClosed(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*semantic.InferencePathV1)
		want   string
	}{
		{
			name: "duplicate record",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[1].RecordID = path.Edges[0].RecordID
			},
			want: "stable-id-collision",
		},
		{
			name: "unknown kind",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[1].Kind = semantic.InferenceKind("UNKNOWN")
			},
			want: "unknown inference kind",
		},
		{
			name: "kind phase mismatch",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[1].Phase.Phase = semantic.PhaseProjection
				path.Edges[1].Authority = semantic.AuthorityBinding{
					Layer: semantic.AuthorityDerived, Effect: semantic.AuthorityProject,
				}
			},
			want: "phase or authority binding",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := clonePath(completeInferenceFixture().path)
			test.mutate(&path)
			assertValidatorError(t, path, test.want)
		})
	}
}

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

func TestClosureOnlyOutcomeCasesAreDependencyLocalNotRun(t *testing.T) {
	cases := []string{
		"missing edge UNKNOWN",
		"zero requirements UNKNOWN",
		"two paths one incomplete UNKNOWN",
		"wrong phase order FAIL_CLOSED",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Skip("dependency-local NOT_RUN: no internal/provenance/pathclosure evaluator exists on origin/dev")
		})
	}
}
