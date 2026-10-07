package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOptionalRecordValueAssemblySynthesizesPresenceAndZeroValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/optional-record-value-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateSourceAssembly(context.Background(), "optional-record-value-assembly.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.RecordAssembly
	if receipt == nil || receipt.Status != "COMPLETE_FINITE" || receipt.SelectedMask != 7 ||
		receipt.Passed != 2 || receipt.Total != 2 || receipt.FieldsPassed != 6 || receipt.FieldsTotal != 6 || receipt.ModelCalls != 0 {
		t.Fatalf("optional value assembly did not select the fully matching candidate: %+v", receipt)
	}
	if len(receipt.Cases) != 2 || !strings.Contains(string(receipt.Cases[0].Actual), `"note":""`) ||
		!strings.Contains(string(receipt.Cases[0].Actual), `"active":false`) || !strings.Contains(string(receipt.Cases[0].Actual), `"total":0`) {
		t.Fatalf("present empty and zero values were not emitted: %+v", receipt.Cases)
	}
	var absent map[string]json.RawMessage
	if err := json.Unmarshal(receipt.Cases[1].Actual, &absent); err != nil || len(absent) != 0 {
		t.Fatalf("absent optional fields were not omitted: actual=%s err=%v", receipt.Cases[1].Actual, err)
	}
	for i, field := range receipt.Cases[0].Fields {
		if field.Presence != "optional" || field.ExpectedPresent == nil || !*field.ExpectedPresent ||
			field.ActualPresent == nil || !*field.ActualPresent || !field.Passed {
			t.Fatalf("case 0 field %d lost explicit presence: %+v", i, field)
		}
	}
	for i, field := range receipt.Cases[1].Fields {
		if field.Presence != "optional" || field.ExpectedPresent == nil || *field.ExpectedPresent ||
			field.ActualPresent == nil || *field.ActualPresent || field.Expected != "ABSENT" || field.Actual != "ABSENT" || !field.Passed {
			t.Fatalf("case 1 field %d lost absence: %+v", i, field)
		}
	}
	if !strings.Contains(result.GoooSource, `picked "note" -> "value_second"`) ||
		!strings.Contains(result.GoooSource, `picked "active" -> "value_second"`) ||
		!strings.Contains(result.GoooSource, `picked "total" -> "value_second"`) {
		t.Fatalf("selected optional field choices were not retained in Gooo: %s", result.GoooSource)
	}
	if scope, ok := result.Report.CompletenessReceipt.Scope["record_body_scope"].(string); !ok || !strings.Contains(scope, "source-declared bounded field synthesis") {
		t.Fatalf("completeness receipt did not describe optional synthesis: %#v", result.Report.CompletenessReceipt.Scope["record_body_scope"])
	}
	modelInput, err := ExportRecordAssemblyContext(context.Background(), "optional-record-value-assembly.gooo", source, "Build", false)
	if err != nil || modelInput.Context == nil {
		t.Fatalf("optional model context export failed: input=%+v err=%v", modelInput, err)
	}
	modelParts := strings.Join(modelInput.Context.Parts, " ")
	if modelInput.Context.Status != "ENCODED" ||
		!strings.Contains(modelParts, `"presence":"optional"`) ||
		!strings.Contains(modelParts, `"type_id":"urn:gooo:type:integer"`) ||
		strings.Contains(modelParts, "discard") || strings.Contains(modelParts, `"amount":42`) {
		t.Fatalf("model context omitted optional field typing or included case outcomes: input=%+v err=%v", modelInput.Context, err)
	}
	for _, claim := range result.Report.CompletenessReceipt.NotClaimed {
		if claim == "source-declared synthesis of optional record values" {
			t.Fatal("completed bounded optional synthesis remains marked as unimplemented")
		}
	}
	replayed, err := RealizeSourceAssembly(context.Background(), "optional-record-value-assembly.gooo", source, result)
	if err != nil || replayed.ModelCalls != 0 || replayed.Source != result.GoooSource {
		t.Fatalf("optional synthesis checkpoint did not replay: replay=%+v err=%v", replayed, err)
	}
}

func TestOptionalRecordValueAssemblyRejectsWrongPointerTypeBeforeInference(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/optional-record-value-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	invalid := []byte(strings.Replace(string(source), `alternative "&amount"`, `alternative "&active"`, 1))
	if err := ValidateSourceAssembly(context.Background(), "optional-record-value-invalid.gooo", invalid, "Build"); err == nil {
		t.Fatal("wrong optional pointer type passed source preflight")
	}
	nullOutput := []byte(strings.Replace(string(source), `\"note\":\"\"`, `\"note\":null`, 1))
	if err := ValidateSourceAssembly(context.Background(), "optional-record-value-null.gooo", nullOutput, "Build"); err == nil {
		t.Fatal("JSON null was accepted as optional absence")
	}
}
