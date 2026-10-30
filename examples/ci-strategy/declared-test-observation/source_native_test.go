package declaredtest

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestAuthoredAssessmentsExecuteNativeAndSavedComposition(t *testing.T) {
	raw, err := os.ReadFile("cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases bodyexecution.CompositionCases
	if err := json.Unmarshal(raw, &cases); err != nil || len(cases.Cases) != 21 {
		t.Fatal("native authored denominator changed", err)
	}
	composition, err := bodyexecution.GenerateCompositionWithOptions(context.Background(), sourcePath, source, cases,
		bodyexecution.CompositionOptions{EntryActivity: "Assess"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(composition)
	if err != nil {
		t.Fatal(err)
	}
	var saved bodyexecution.Composition
	if err := json.Unmarshal(encoded, &saved); err != nil {
		t.Fatal(err)
	}
	run, err := bodyexecution.ExecuteComposition(context.Background(), sourcePath, source, saved, cases,
		filepath.Join(runtime.GOROOT(), "bin", "go"))
	if err != nil || run.FinitePassed != 21 || run.FiniteTotal != 21 ||
		!run.RuntimeReplayed || !run.ProjectionReplayed || run.ModelCalls != 0 {
		t.Fatal("native/saved source assessment failed", err, run.Stage, run.FinitePassed, run.FiniteTotal)
	}
}

func TestSourceReaderRetainsNestedObservationIdentity(t *testing.T) {
	path, err := filepath.Abs("main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "run", "../../record-origin-reader",
		"--activity", "Assess", "--field", "urn:declared-tests:assessment/observation-source", path)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("original Gooo reader: %v\n%s", err, raw)
	}
	var query struct {
		Steps  int
		Budget int `json:"step_budget"`
		Result struct{ Status, Paths string }
	}
	if err := json.Unmarshal(raw, &query); err != nil {
		t.Fatal(err)
	}
	want := "input:urn:declared-tests:trial/observation > urn:declared-tests:observation/source;"
	if query.Result.Status != "RESOLVED" || query.Result.Paths != want || query.Steps < 1 || query.Steps > query.Budget {
		t.Fatal("observation's subject merged into the plan identity", query)
	}
}
