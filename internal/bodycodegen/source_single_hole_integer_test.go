package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSourceSingleHoleIntegerLoop(t *testing.T) {
	source := []byte("package threshold\nnamespace threshold\nentity Integer id \"threshold://integer\"\n" +
		"activity Threshold(Integer) -> Integer computes `let offset = 0\n" +
		"for offset = 0; offset < 8; offset = offset + 1 {\n" +
		"if __GOOO_BODY_HOLE_stop__ { return offset }\n}\nreturn -1` assembling {\n" +
		"source_fill intent \"Find the first counter at least as large as the input.\" {\n" +
		"hole \"stop\"\ncandidate \"atleast\" { fill \"stop\" \"offset >= input\" }\n" +
		"candidate \"beyond\" { fill \"stop\" \"offset > input\" }\n}\n" +
		"case \"-1\" -> \"0\"\ncase \"3\" -> \"3\"\n" +
		"holdout_case \"7\" -> \"7\"\nholdout_case \"8\" -> \"-1\"\n}\n")
	file, diagnostics := ParseBodyFile("single-loop-fill.gooo", source)
	if file == nil || diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != "Threshold" {
			continue
		}
		result, err := GenerateWithSourceIRBodyFill(context.Background(), "single-loop-fill.gooo", source,
			"Threshold", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
		if err != nil {
			t.Fatal(err)
		}
		r := result.Report.BodyFill
		if r == nil || r.SelectedCandidateID != "atleast" || r.TestCasesPassed != 2 || r.HoldoutCasesPassed != 2 ||
			r.LocalModelPredictions != 0 || r.ExternalProviderCalls != 0 || !result.Report.TypecheckPassed ||
			strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(result.GoooSource, "source_fill") {
			t.Fatalf("single source-owned integer choice was not checked: %+v", r)
		}
		return
	}
	t.Fatal("threshold activity missing")
}
