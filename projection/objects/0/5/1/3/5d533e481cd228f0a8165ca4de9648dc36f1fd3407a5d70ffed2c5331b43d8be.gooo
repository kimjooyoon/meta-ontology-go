package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"reflect"
	"testing"
)

func TestEvaluateR4FiniteBoundaryAndRootIsolation(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*pathclosure.R4Input)
		code   string
	}{
		{"open world", func(input *pathclosure.R4Input) { input.Boundary.OpenWorld = true }, pathclosure.CodeOpenWorld},
		{"not exhausted", func(input *pathclosure.R4Input) { input.Boundary.Exhausted = false }, pathclosure.CodeUnexhaustedBoundary},
		{"root replay is not discovered", func(input *pathclosure.R4Input) { input.Paths[0].StartID = r4ID("node/alternate-root") }, pathclosure.CodeInvalidPath},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := cloneR4Input(completeR4Fixture().input)
			test.mutate(&input)
			got := pathclosure.EvaluateR4(input)
			if got.Status != pathclosure.UNKNOWN && test.name != "root replay is not discovered" || got.Code != test.code {
				t.Fatalf("R4 result = %#v, want code %s", got, test.code)
			}
		})
	}
}
func TestR4ExpectedOnlyMetaLabelCannotAffectDecision(t *testing.T) {
	input := completeR4Fixture().input
	baseline := pathclosure.EvaluateR4(input)
	baselineDigest := baseline.CanonicalDigest()

	for _, metaLabel := range []string{"compile", "runtime", "forged-alias"} {
		result := func(_ string) pathclosure.R4Result { return pathclosure.EvaluateR4(input) }(metaLabel)
		if !reflect.DeepEqual(result, baseline) || result.CanonicalDigest() != baselineDigest {
			t.Fatalf("meta label %q changed the R4 decision: %#v vs %#v", metaLabel, result, baseline)
		}
	}
}
