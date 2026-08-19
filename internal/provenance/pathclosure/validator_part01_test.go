package pathclosure_test

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
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
