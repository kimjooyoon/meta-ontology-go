package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func recordAssemblyFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestRecordAssemblySelectsTypedFieldsAndReplaysCheckpoint(t *testing.T) {
	source := recordAssemblyFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "record.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.Status != "COMPLETE_FINITE" || r.Passed != 5 || r.Total != 5 || r.FieldsPassed != 15 || r.FieldsTotal != 15 ||
		r.SelectedMask != 7 || len(r.Attempts) != 8 || r.ModelCalls != 0 || !strings.Contains(result.GoooSource, "picked \"title\" -> \"value_second\"") {
		t.Fatalf("record selection differs: %+v", r)
	}
	realized, err := RealizeSourceAssembly(context.Background(), "record.gooo", source, result)
	if err != nil || realized.ModelCalls != 0 || realized.Source != result.GoooSource {
		t.Fatal("record realization", err)
	}
	next, err := g.GenerateSourceAssembly(context.Background(), "record.gooo", []byte(result.GoooSource), "Select")
	if err != nil || next.Source != result.Source || next.GoooSource != result.GoooSource {
		t.Fatal("record checkpoint fixed point", err)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "record.gooo", []byte(result.GoooSource), next); err != nil {
		t.Fatal(err)
	}
}

func TestBooleanRecordFieldAssemblyPreservesTypedValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/boolean-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateSourceAssembly(context.Background(), "boolean-record.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.RecordAssembly
	if receipt == nil || receipt.Status != "COMPLETE_FINITE" || receipt.SelectedMask != 7 ||
		receipt.Passed != 4 || receipt.Total != 4 || receipt.FieldsPassed != 12 || receipt.FieldsTotal != 12 || receipt.ModelCalls != 0 {
		t.Fatalf("Boolean record assembly did not complete the declared cases: %+v", receipt)
	}
	booleanType := "urn:gooo:type:boolean"
	for _, record := range result.Report.RecordTypes {
		if record.Name != "Gate" {
			continue
		}
		for _, field := range record.Fields {
			if field.Name == "enabled" || field.Name == "reviewed" {
				if field.TypeID != booleanType {
					t.Fatalf("field %s lost Boolean type: %+v", field.Name, field)
				}
			}
		}
	}
	for index, observation := range receipt.Cases {
		var actual map[string]any
		if err := json.Unmarshal(observation.Actual, &actual); err != nil {
			t.Fatal(err)
		}
		if _, ok := actual["enabled"].(bool); !ok {
			t.Fatalf("case %d encoded enabled as %T: %s", index, actual["enabled"], observation.Actual)
		}
		if _, ok := actual["reviewed"].(bool); !ok {
			t.Fatalf("case %d encoded reviewed as %T: %s", index, actual["reviewed"], observation.Actual)
		}
	}
	if !strings.Contains(result.GoooSource, `picked "enabled" -> "value_second"`) ||
		!strings.Contains(result.Source, "bool") {
		t.Fatalf("selected source does not retain Boolean alternatives: %s", result.GoooSource)
	}
	replayed, err := RealizeSourceAssembly(context.Background(), "boolean-record.gooo", source, result)
	if err != nil || replayed.ModelCalls != 0 || replayed.Source != result.GoooSource {
		t.Fatalf("Boolean assembly checkpoint replay differs: replay=%+v err=%v", replayed, err)
	}
	bad := []byte(strings.Replace(string(source), `alternative "input0"`, `alternative "\"bad\""`, 1))
	if err := ValidateSourceAssembly(context.Background(), "boolean-record-bad.gooo", bad, "Build"); err == nil {
		t.Fatal("string expression accepted for Boolean record field")
	}
}

func TestIntegerRecordFieldAssemblyPreservesTypedValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/integer-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateSourceAssembly(context.Background(), "integer-record.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.RecordAssembly
	if receipt == nil || receipt.Status != "COMPLETE_FINITE" || receipt.SelectedMask != 7 ||
		receipt.Passed != 3 || receipt.Total != 3 || receipt.FieldsPassed != 9 || receipt.FieldsTotal != 9 || receipt.ModelCalls != 0 {
		t.Fatalf("integer record assembly did not complete the declared cases: %+v", receipt)
	}
	integerType := "urn:gooo:type:integer"
	for _, record := range result.Report.RecordTypes {
		if record.Name != "Stats" {
			continue
		}
		for _, field := range record.Fields {
			if field.Name == "total" && field.TypeID != integerType {
				t.Fatalf("integer field lost semantic type: %+v", field)
			}
		}
	}
	for index, observation := range receipt.Cases {
		var actual map[string]any
		if err := json.Unmarshal(observation.Actual, &actual); err != nil {
			t.Fatal(err)
		}
		if actual["total"] != float64([]int{1, 0, 42}[index]) {
			t.Fatalf("case %d encoded total incorrectly: %s", index, observation.Actual)
		}
	}
	if !strings.Contains(result.Source, "int64") || !strings.Contains(result.GoooSource, `picked "total" -> "value_second"`) {
		t.Fatalf("selected source does not retain integer field semantics: %s", result.GoooSource)
	}
	replayed, err := RealizeSourceAssembly(context.Background(), "integer-record.gooo", source, result)
	if err != nil || replayed.ModelCalls != 0 || replayed.Source != result.GoooSource {
		t.Fatalf("integer assembly checkpoint replay differs: replay=%+v err=%v", replayed, err)
	}
	bad := []byte(strings.Replace(string(source), `alternative "input0 + 1"`, `alternative "\"wrong\""`, 1))
	if err := ValidateSourceAssembly(context.Background(), "integer-record-bad.gooo", bad, "Build"); err == nil {
		t.Fatal("string expression accepted for integer record field")
	}
}

func TestRecordAssemblyPartialScoreAndSourcePreflight(t *testing.T) {
	source := recordAssemblyFixture(t)
	g, _ := NewTypedPathGenerator("")
	partial := []byte(strings.Replace(string(source), "attempts \"8\"", "attempts \"1\"", 1))
	r, err := g.GenerateSourceAssembly(context.Background(), "partial.gooo", partial, "Select")
	if err != nil {
		t.Fatal(err)
	}
	p := r.Report.RecordAssembly
	if p.Status != "PARTIAL_FINITE" || p.Passed != 2 || p.Total != 5 || p.FieldsPassed != 6 || p.FieldsTotal != 15 || len(p.Attempts) != 1 {
		t.Fatal("partial observations", p)
	}
	if err := completeness.Validate(r.Report.CompletenessReceipt); err != nil {
		t.Fatal("partial common record receipt", err)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "partial.gooo", partial, r); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"alternative \"input0.title\"", "alternative \"input1\""},
		{"alternative \"input0.title\"", "alternative \"missing.title\""},
		{"field_value at \"1\"", "field_value at \"0\""},
		{"\\\"queued\\\"", "null"}} {
		bad := []byte(strings.Replace(string(source), pair[0], pair[1], 1))
		if err = ValidateSourceAssembly(context.Background(), "bad.gooo", bad, "Select"); err == nil {
			t.Fatal("invalid field contract accepted", pair)
		}
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = g.GenerateSourceAssembly(cancelled, "record.gooo", source, "Select"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation", err)
	}
}

func TestRecordAssemblyImmutableModelConcurrencyAndCompleteContext(t *testing.T) {
	source := recordAssemblyFixture(t)
	g, err := NewTypedPathGenerator(writeThreeContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			result, err := g.GenerateSourceAssembly(context.Background(), "record.gooo", source, "Select")
			if err != nil {
				t.Error(err)
				return
			}
			r := result.Report.RecordAssembly
			if r.ModelCalls != 1 || r.PredictNS < 1 || r.Prediction == nil || r.Context.Status != "ENCODED" || r.SelectedMask != 7 || r.Model.ModelSchema == "" {
				t.Errorf("local ordinal prediction: calls=%d context=%+v", r.ModelCalls, r.Context)
			}
			if strings.Contains(r.Context.Text, "value_case") || !strings.Contains(r.Context.Text, "Keep the original title.") {
				t.Error("context lost intent or included cases")
			}
			if _, err = RealizeSourceAssembly(context.Background(), "record.gooo", source, result); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	large := []byte(strings.Replace(string(source), "Keep the original title.", strings.Repeat("한", 120), 1))
	result, err := g.GenerateSourceAssembly(context.Background(), "large.gooo", large, "Select")
	if err != nil || result.Report.RecordAssembly.Context.Status != "DECLINED_TO_DETERMINISTIC" || result.Report.RecordAssembly.ModelCalls != 0 {
		t.Fatal("complete input decline", err)
	}
}

func TestRecordAssemblyReplayRejectsMutatedFiniteAndTypedObservations(t *testing.T) {
	source := recordAssemblyFixture(t)
	g, _ := NewTypedPathGenerator("")
	r, err := g.GenerateSourceAssembly(context.Background(), "record.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Result){
		func(r *Result) { r.Report.RecordAssembly.Cases[0].Fields[0].Actual = "wrong" },
		func(r *Result) { r.Report.RecordAssembly.FieldsPassed-- },
		func(r *Result) { r.Report.RecordAssembly.SelectedMask = 0 },
		func(r *Result) { r.Report.RecordAssembly.Choices[0].FieldID = "changed" },
		func(r *Result) { r.Report.RecordAssembly.Ranking[0] = 7 },
		func(r *Result) { r.Report.RecordTypes[0].Fields[0].ID = "changed" },
		func(r *Result) { r.GoooSource += "\n" },
		func(r *Result) { r.Report.CompletenessReceipt.Scope["record_assembly"] = map[string]any{} },
		func(r *Result) {
			for i := range r.Report.CompletenessReceipt.Dimensions {
				if r.Report.CompletenessReceipt.Dimensions[i].ID == "declared_record_field_accuracy" {
					r.Report.CompletenessReceipt.Dimensions[i].Unit = "unrelated counts"
				}
			}
		},
	} {
		raw, _ := json.Marshal(r)
		var changed Result
		if err := json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		if _, err = RealizeSourceAssembly(context.Background(), "record.gooo", source, changed); err == nil {
			t.Fatal("mutated record replay accepted")
		}
	}
	if !reflect.DeepEqual(r.Report.RecordAssembly.Cases[0].Actual, json.RawMessage(`{"reason":"검토:accepted","state":"ready","title":"한글"}`)) {
		t.Fatal("record case actual", string(r.Report.RecordAssembly.Cases[0].Actual))
	}
}
