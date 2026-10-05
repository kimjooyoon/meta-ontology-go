package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGenerateWithSourceIRBodyFillUsesDeclaredCandidatesDeterministically(t *testing.T) {
	source := []byte("package sample\nnamespace sample\nentity Integer id \"sample://integer\"\n" +
		"activity Lift(Integer) -> Integer computes `let base = __GOOO_BODY_HOLE_seed__\n" +
		"let increment = __GOOO_BODY_HOLE_step__\nreturn base + increment` assembling {\n" +
		"source_fill intent \"Compose base and increment\" {\n" +
		"hole \"seed\"\nhole \"step\"\n" +
		"candidate \"add_one\" { fill \"seed\" \"input + 0\" fill \"step\" \"1\" }\n" +
		"candidate \"double\" { fill \"seed\" \"input * 2\" fill \"step\" \"0\" }\n" +
		"}\ncase \"0\" -> \"1\"\ncase \"2\" -> \"3\"\n}\n")
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "source-fill.gooo", source, "Lift", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.SelectedCandidateID != "add_one" || result.Report.BodyFill.FunctionalAccuracyPct != 100 {
		t.Fatalf("source fill did not produce a finite measured result: %+v", result.Report.BodyFill)
	}
	if strings.Contains(result.GoooSource, "source_fill") || strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatalf("generated Gooo source retained assembly instructions or holes:\n%s", result.GoooSource)
	}
	if !strings.Contains(result.Source, "base = (input + 0)") || !strings.Contains(result.Source, "increment int64 = 1") || !result.Report.TypecheckPassed {
		t.Fatalf("selected assignments were not compiled into the generated body: %s", result.Source)
	}
	replay, replayDiagnostics := syntax.Parse(result.GoooSource)
	if replayDiagnostics.HasErrors() {
		t.Fatalf("generated Gooo source cannot be replayed: %v\n%s", replayDiagnostics, result.GoooSource)
	}
	if replay.Declarations[1].(*syntax.ActivityDecl).Assembly != nil {
		t.Fatal("replay source still requires a model choice")
	}
}
