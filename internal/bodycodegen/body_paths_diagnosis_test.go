package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func diagnosisFixture(t *testing.T) ([]byte, pathplan.Document) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/path-diagnosis.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/path-diagnosis-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	document, err := pathplan.DecodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, document
}

func TestTypedPathDiagnosisRetainsNativeBodyAndUnresolvedIntent(t *testing.T) {
	source, document := diagnosisFixture(t)
	baseline, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Probe", document, "")
	if err != nil {
		t.Fatal(err)
	}
	options := TypedPathOptions{Diagnosis: &PathDiagnosisOptions{Inputs: []int64{2, 3}, MaxCandidates: 2}}
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", document, "", options)
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyPaths
	if result.Source != baseline.Source || receipt.Search.Selection.ModelCalls != 0 || result.Report.RepositoryWrites != 0 ||
		receipt.FunctionalCompleteness != 100 || receipt.Diagnosis == nil || receipt.Diagnosis.CaseIndistinguishable != 2 ||
		receipt.Diagnosis.ProbeDistinguished != 1 || receipt.Diagnosis.ProbeUnresolved != 1 || receipt.Diagnosis.ModelPredictions != 0 ||
		receipt.DiagnosisBudget != 2 || len(receipt.DiagnosisOptionsSHA256) != 71 ||
		!strings.HasPrefix(receipt.DiagnosisOptionsSHA256, "sha256:") || receipt.Timing.DiagnosisMS <= 0 {
		t.Fatalf("native diagnosis lost finite uncertainty: %+v", receipt)
	}
	witness := receipt.Diagnosis.Candidates[1].Witness
	if witness == nil || witness.Input != 3 || witness.Reference != 1 || witness.Alternative != -1 {
		t.Fatal("diagnostic comparison was turned into an expectation")
	}
	raw, _ := json.Marshal(baseline.Report.BodyPaths)
	if strings.Contains(string(raw), `"diagnosis`) {
		t.Fatal("default receipt gained diagnostic fields")
	}
	generator, _ := NewTypedPathGenerator("")
	retained, err := generator.Generate(context.Background(), "fixture.gooo", source, "Probe", document, options)
	if err != nil || retained.Source != result.Source || retained.Report.BodyPaths.Diagnosis.ProbeDistinguished != 1 {
		t.Fatal("retained path options did not reach native diagnosis", err)
	}
}

func TestTypedPathDiagnosisRejectsInvalidOptionsBeforeModelLoad(t *testing.T) {
	source, document := diagnosisFixture(t)
	for _, options := range []*PathDiagnosisOptions{
		{Inputs: []int64{3}, MaxCandidates: 0}, {Inputs: nil, MaxCandidates: 2},
		{Inputs: make([]int64, 33), MaxCandidates: 2}, {Inputs: []int64{3}, MaxCandidates: 65},
	} {
		_, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", document,
			"absent-model.json", TypedPathOptions{Diagnosis: options})
		if err == nil || !strings.Contains(err.Error(), "diagnosis") {
			t.Fatal("invalid options reached model loading", err)
		}
	}
}
