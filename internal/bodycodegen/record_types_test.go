package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"
)

func recordFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/native-records.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestRecordBodyConstructionSelectorsAndStableNativeNames(t *testing.T) {
	source := recordFixture(t)
	for _, activity := range []string{"Propose", "ReviewCandidate", "Label", "Echo", "Same"} {
		result, err := Generate("records.gooo", source, activity)
		if err != nil || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay ||
			!result.Report.RouteEquivalence.Equivalent || result.Report.CompletenessPercent != 100 {
			t.Fatalf("%s: %v %+v", activity, err, result.Report)
		}
		if len(result.Report.RecordTypes) == 0 || !strings.Contains(result.Source, "GoooRecord") ||
			!strings.Contains(result.Source, "GoooField") {
			t.Fatal("record identities or native value layout missing", activity)
		}
	}
	scalar, err := Generate("records.gooo", source, "Score")
	if err != nil || len(scalar.Report.RecordTypes) != 0 || strings.Contains(scalar.Source, "type GoooRecord") {
		t.Fatalf("scalar assembly projection gained unrelated declarations: %v", err)
	}
}

func TestRecordBodyRejectsIncompleteAndUnsupportedValues(t *testing.T) {
	source := string(recordFixture(t))
	old := `Candidate{title: input1, state: "ready"}`
	for _, body := range []string{
		`Candidate{title: input1}`, `Candidate{title: input1, state: "ready", other: "x"}`,
		`Candidate{title: input1, title: "again", state: "ready"}`, `Candidate{input1, "ready"}`,
		`Candidate{title: input1, state: 1}`, `Other{title: input1, state: "ready"}`,
		`[]string{"x"}`, `struct{ title string }{title: input1}`,
	} {
		candidate := strings.Replace(source, old, body, 1)
		if _, err := Generate("bad.gooo", []byte(candidate), "Propose"); err == nil {
			t.Fatal("accepted unsupported construction", body)
		}
	}
	for _, body := range []string{`input.state = "changed"; return input`, `return input.missing`, `return input.title()`} {
		candidate := strings.Replace(source, `computes "return input"`, "computes `"+body+"`", 1)
		if _, err := Generate("bad.gooo", []byte(candidate), "Echo"); err == nil {
			t.Fatal("accepted unsupported record use", body)
		}
	}
}

func TestRecordProjectionSupportsEveryEquivalentConditionalRoute(t *testing.T) {
	source := recordFixture(t)
	base, err := Generate("records.gooo", source, "Propose")
	if err != nil {
		t.Fatal(err)
	}
	body := `if input0 >= 0 { return Candidate{title: input1, state: "ready"} } else { return Candidate{title: input1, state: "wait"} }`
	parameters := []InputParameter{{Name: "input0", Type: "int64"}, {Name: "input1", Type: "string"}}
	for _, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
		result, err := generateRouteParameters("records", "Propose", base.Report.ActivityID, parameters,
			"Candidate", body, route, base.Report.RecordTypes...)
		if err != nil || !result.report.RouteEquivalence.Equivalent || !result.report.TypecheckPassed {
			t.Fatalf("%s: %v", route, err)
		}
	}
}

func TestBooleanRecordFieldIsGeneratedAndFiniteCasesAreTyped(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/boolean-records.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateSourceAssembly(context.Background(), "boolean.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.RecordAssembly
	if receipt == nil || receipt.Status != "COMPLETE_FINITE" || receipt.Passed != 2 || receipt.Total != 2 ||
		!result.Report.TypecheckPassed || !strings.Contains(result.Source, " bool `json:\"enabled\"`") {
		t.Fatalf("Boolean record field was not generated and observed as Boolean: receipt=%+v source=%s", receipt, result.Source)
	}
	if len(receipt.Cases) != 2 || !strings.Contains(string(receipt.Cases[0].Actual), `"enabled":true`) ||
		!strings.Contains(string(receipt.Cases[1].Actual), `"enabled":false`) {
		t.Fatalf("Boolean observations were not JSON booleans: %+v", receipt.Cases)
	}
	if len(receipt.Cases[0].Fields) != 1 || receipt.Cases[0].Fields[0].TypeID != "urn:gooo:type:boolean" {
		t.Fatalf("field observation lost its semantic type: %+v", receipt.Cases[0].Fields)
	}
}
