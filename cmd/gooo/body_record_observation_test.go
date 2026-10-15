package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordObserveCLIReportsAcquiredCasesAndOracleFailure(t *testing.T) {
	const source = `package question
namespace question
entity Integer id "question://integer"
entity Box id "question://box" fields { field value id "question://value" type integer required one }
activity Choose(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "0 - input" intent "Negate."
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[17]" oracle "Expected" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
activity Expected(Integer) -> Box computes "return Box{value: 0 - input}"`
	filename := filepath.Join(t.TempDir(), "questions.gooo")
	if err := os.WriteFile(filename, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", filename}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.Observation == nil || r.Total != 2 || r.Passed != 2 || r.ModelCalls != 0 {
		t.Fatal(r)
	}
	for _, format := range []string{"text", "markdown"} {
		out.Reset()
		output := bodyConstructOutput{Construction: bodyexecution.JointConstruction{Initial: bodyexecution.Composition{
			Preparations: []bodyexecution.CompositionStep{{Generation: result}}}}}
		if err := writeBodyConstructResult(&out, output, format); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "1 initial + 1 acquired = 2 effective") || !strings.Contains(out.String(), "1 surviving; 0 unobserved; 0 unavailable") {
			t.Fatal(out.String())
		}
	}
	bad := strings.Replace(source, "return Box{value: 0 - input}", "return Box{value: 1 - input}", 1)
	if err := os.WriteFile(filename, []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", filename}, &out, &diagnostics); code != exitFailure {
		t.Fatal(code, out.String())
	}
	var failure struct {
		Record *bodycodegen.RecordAssemblyReceipt `json:"record_assembly"`
	}
	if err := json.Unmarshal(out.Bytes(), &failure); err != nil || failure.Record == nil || failure.Record.Observation.Status != "FAILED" || failure.Record.Observation.OracleInitialCases[0].Passed {
		t.Fatal(err, out.String())
	}
}
