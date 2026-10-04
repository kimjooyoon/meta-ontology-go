package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordSourceAssemblyCLIExportsAndConstructsWithoutExternalPlan(t *testing.T) {
	source := "../../examples/body-codegen/record-field-assembly.gooo.fixture"
	var output, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Select", source}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil || result.Report.RecordAssembly == nil || result.Report.RecordAssembly.FieldsPassed != 15 {
		t.Fatal("record CLI result", err)
	}
	output.Reset()
	if code := run([]string{"body-context", "--include-plan", "--activity", "Select", source}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var exported bodycodegen.RecordAssemblyContextExport
	if err := json.Unmarshal(output.Bytes(), &exported); err != nil || exported.ExpandedPlan == nil || exported.Context.Status != "ENCODED" ||
		exported.ModelPredictions != 0 || exported.CandidateTests != 0 || exported.ContractSHA256 != result.Report.RecordAssembly.ContractSHA256 {
		t.Fatal("record CLI context", err, string(output.Bytes()))
	}
}
