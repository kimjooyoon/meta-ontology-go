package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
