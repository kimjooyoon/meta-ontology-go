package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestCompositionSynthesizesOptionalRecordValuesAndExecutesThem(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/optional-record-value-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/optional-record-value-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "optional-synthesis.gooo", source, suite, "")
	if err != nil {
		t.Fatal("optional synthesis generation", err)
	}
	assembly := prior.Steps[0].Generation.Report.RecordAssembly
	if assembly == nil || assembly.Status != "COMPLETE_FINITE" || assembly.SelectedMask != 7 ||
		assembly.Passed != 2 || assembly.Total != 2 || assembly.FieldsPassed != 6 || assembly.FieldsTotal != 6 {
		t.Fatalf("source assembly did not select the typed optional values: %+v", assembly)
	}
	run, err := ExecuteComposition(context.Background(), "optional-synthesis.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || !run.RuntimeReplayed {
		t.Fatalf("generated optional values did not execute and replay: err=%v run=%+v", err, run)
	}
	for caseIndex, trace := range run.Traces {
		if len(trace.Deliveries) != 2 {
			t.Fatalf("case %d activity deliveries=%d", caseIndex, len(trace.Deliveries))
		}
		build, echo := trace.Deliveries[0], trace.Deliveries[1]
		if !bytes.Equal(build.Actual, echo.Input) || !bytes.Equal(build.Actual, echo.Actual) {
			t.Fatalf("case %d did not carry the synthesized optional record: %+v", caseIndex, trace.Deliveries)
		}
		for _, fields := range [][]CompositionRecordField{build.ActualFields, echo.InputFields, echo.ActualFields} {
			if len(fields) != 3 {
				t.Fatalf("case %d optional field observations=%d", caseIndex, len(fields))
			}
			for _, field := range fields {
				if field.Presence != "optional" || field.Present == nil {
					t.Fatalf("case %d lost optional presence: %+v", caseIndex, field)
				}
				wantPresent := caseIndex == 0
				if *field.Present != wantPresent {
					t.Fatalf("case %d field %q presence=%t", caseIndex, field.Name, *field.Present)
				}
			}
		}
		if caseIndex == 0 {
			want := map[string]string{"note": `""`, "active": "false", "total": "0"}
			for _, field := range build.ActualFields {
				if string(field.Value) != want[field.Name] {
					t.Fatalf("present zero field %q = %s, want %s", field.Name, field.Value, want[field.Name])
				}
			}
		} else {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(build.Actual, &object); err != nil || len(object) != 0 {
				t.Fatalf("absent values were emitted as JSON fields: actual=%s err=%v", build.Actual, err)
			}
			for _, field := range build.ActualFields {
				if field.Value != nil {
					t.Fatalf("absent field %q retained a value: %s", field.Name, field.Value)
				}
			}
		}
	}
}
