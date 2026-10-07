package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func nativeRecordFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/native-records.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/native-records-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionNativeRecordsAndActualFieldIdentities(t *testing.T) {
	source, suite := nativeRecordFixture(t)
	prior, err := GenerateComposition(context.Background(), "records.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Plan.Records) != 2 || len(prior.Steps[0].Generation.Report.RecordTypes) != 0 {
		t.Fatal("record layouts missing or scalar assembly gained declarations")
	}
	for _, record := range prior.Plan.Records {
		for _, field := range record.Fields {
			if field.Presence != "" {
				t.Fatalf("required V3 field gained optional metadata: %+v", field)
			}
		}
	}
	run, err := ExecuteComposition(context.Background(), "records.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 36 || run.FiniteTotal != 36 || !run.RuntimeReplayed || run.ModelCalls != 0 {
		t.Fatalf("native records: %v %+v", err, run)
	}
	fields := 0
	for _, trace := range run.Traces {
		for i, delivery := range trace.Deliveries {
			node := prior.Plan.Activities[i]
			fields += len(delivery.ActualFields) + len(delivery.InputFields)
			for p, input := range delivery.Inputs {
				fields += len(input.Fields)
				if input.ProducerID != "" {
					from := node.Inputs[p].From
					if !bytes.Equal(input.Value, trace.Deliveries[from].Actual) {
						t.Fatal("record input differs from its actual producer")
					}
				}
			}
			for _, field := range delivery.ActualFields {
				if field.ID == "" || field.Name == "" || field.Value == nil {
					t.Fatal("actual field identity/value missing")
				}
			}
		}
	}
	if fields != 96 {
		t.Fatalf("field observations=%d, want96", fields)
	}
	suite.Cases[0].Expected["Propose"] = json.RawMessage(`{"title":"gooo","state":"wait"}`)
	partial, err := ExecuteComposition(context.Background(), "records.gooo", source, prior, suite, nativeTool())
	if err != nil || partial.FinitePassed != 35 || partial.FiniteTotal != 36 {
		t.Fatalf("record expectation partial: %v %+v", err, partial)
	}
}

func TestCompositionRecordMetadataMustReplay(t *testing.T) {
	source, suite := nativeRecordFixture(t)
	prior, err := GenerateComposition(context.Background(), "records.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	for i := range prior.Steps {
		if len(prior.Steps[i].Generation.Report.RecordTypes) > 0 {
			prior.Steps[i].Generation.Report.RecordTypes[0].Fields[0].ID = "records://changed-field"
			break
		}
	}
	if _, err := ExecuteComposition(context.Background(), "records.gooo", source, prior, suite, nativeTool()); err == nil {
		t.Fatal("altered field identities replayed")
	}
}

func TestCompositionRecordObjectsAreCompleteAndExact(t *testing.T) {
	source, suite := nativeRecordFixture(t)
	graph, err := prepareCompositionGraph(context.Background(), "records.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`null`, `[]`, `{}`, `{"title":"x"}`, `{"title":"x","state":null}`,
		`{"title":"x","state":1}`, `{"title":"x","state":"ok","extra":"x"}`,
		`{"Title":"x","state":"ok"}`, `{"title":"x","state":"ok","title":"again"}`,
		`{"title":{},"state":"ok"}`, `{"title":"` + strings.Repeat("x", 1025) + `","state":"ok"}`} {
		if _, err := graph.canonicalValue([]byte(raw), "Candidate"); err == nil {
			t.Fatal("accepted malformed record", raw)
		}
	}
	left, err := graph.canonicalValue([]byte(`{"state":"ready","title":"gooo"}`), "Candidate")
	if err != nil {
		t.Fatal(err)
	}
	right, err := graph.canonicalValue(suite.Cases[0].Expected["Propose"], "Candidate")
	if err != nil || !bytes.Equal(left, right) {
		t.Fatal("JSON field presentation order changed record value", err)
	}
}

func TestCompositionTransportsBooleanRecordFieldsAsJSONBooleans(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/boolean-record-composition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/boolean-record-composition-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "boolean-records.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Plan.Records) != 1 || len(prior.Plan.Records[0].Fields) != 1 ||
		prior.Plan.Records[0].Fields[0].TypeID != "urn:gooo:type:boolean" {
		t.Fatalf("record type was not retained: %+v", prior.Plan.Records)
	}
	graph, err := prepareCompositionGraph(context.Background(), "boolean-records.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{`{"enabled":"true"}`, `{"enabled":null}`, `{"enabled":1}`, `{}`, `{"enabled":true,"extra":false}`} {
		if _, err := graph.canonicalValue([]byte(invalid), "Gate"); err == nil {
			t.Fatalf("accepted invalid Boolean record: %s", invalid)
		}
	}
	invalidSuite := suite
	invalidSuite.Cases = append([]CompositionCase(nil), suite.Cases...)
	invalidSuite.Cases[0].Inputs = map[string]json.RawMessage{"Build": json.RawMessage(`"true"`)}
	if _, err := GenerateComposition(context.Background(), "boolean-records.gooo", source, invalidSuite, "missing-model.json"); err == nil || strings.Contains(err.Error(), "missing-model.json") {
		t.Fatalf("invalid Boolean type reached model loading: %v", err)
	}
	run, err := ExecuteComposition(context.Background(), "boolean-records.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || !run.RuntimeReplayed || run.ModelCalls != 0 {
		t.Fatalf("Boolean record composition: %v %+v", err, run)
	}
	for caseIndex, trace := range run.Traces {
		if len(trace.Deliveries) != 2 {
			t.Fatalf("case %d deliveries=%d", caseIndex, len(trace.Deliveries))
		}
		expectedBuild, err := graph.canonicalValue(suite.Cases[caseIndex].Expected["Build"], "Gate")
		if err != nil {
			t.Fatal(err)
		}
		expectedRelay, err := graph.canonicalValue(suite.Cases[caseIndex].Expected["Relay"], "Gate")
		if err != nil {
			t.Fatal(err)
		}
		if string(trace.Deliveries[0].Actual) != string(expectedBuild) ||
			string(trace.Deliveries[1].Input) != string(trace.Deliveries[0].Actual) ||
			string(trace.Deliveries[1].Actual) != string(expectedRelay) {
			t.Fatalf("case %d failed to deliver typed Boolean values: %+v", caseIndex, trace.Deliveries)
		}
		want := "true"
		if caseIndex == 1 {
			want = "false"
		}
		if len(trace.Deliveries[0].ActualFields) != 1 || string(trace.Deliveries[0].ActualFields[0].Value) != want ||
			len(trace.Deliveries[1].InputFields) != 1 || string(trace.Deliveries[1].InputFields[0].Value) != string(trace.Deliveries[0].ActualFields[0].Value) {
			t.Fatalf("case %d lost Boolean field observations: %+v", caseIndex, trace.Deliveries)
		}
	}
}

func TestCompositionTransportsIntegerRecordFieldsExactly(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/integer-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build.input0": json.RawMessage("9007199254740993"), "Build.input1": json.RawMessage("true"), "Build.input2": json.RawMessage(`"exact"`)},
			Expected: map[string]json.RawMessage{
				"Build": json.RawMessage(`{"total":9007199254740994,"enabled":true,"label":"exact"}`),
				"Echo":  json.RawMessage(`{"total":9007199254740995,"enabled":true,"label":"exact"}`),
			}},
	}}
	prior, err := GenerateComposition(context.Background(), "integer.gooo", source, suite, "")
	if err != nil {
		t.Fatal("integer composition generation", err)
	}
	run, err := ExecuteComposition(context.Background(), "integer.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || run.FiniteTotal != 2 || !run.RuntimeReplayed {
		t.Fatalf("integer composition was not exact and replayable: run=%+v err=%v", run, err)
	}
	if len(run.Traces) != 1 || len(run.Traces[0].Deliveries) != 2 ||
		string(run.Traces[0].Deliveries[1].Actual) != `{"enabled":true,"label":"exact","total":9007199254740995}` {
		t.Fatalf("integer runtime value changed: %+v", run.Traces)
	}
}

func TestCompositionOptionalRecordTransportPreservesAbsenceAndZeroValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/optional-record-transport.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/optional-record-transport-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "optional-records.gooo", source, suite, "")
	if err != nil {
		t.Fatal("optional record composition generation", err)
	}
	if len(prior.Plan.Records) != 1 || len(prior.Plan.Records[0].Fields) != 3 {
		t.Fatalf("optional record layout missing: %+v", prior.Plan.Records)
	}
	for _, field := range prior.Plan.Records[0].Fields {
		if field.Presence != "optional" {
			t.Fatalf("field presence was not retained: %+v", field)
		}
	}
	receiptScope, err := json.Marshal(prior.Steps[0].Generation.Report.CompletenessReceipt.Scope)
	if err != nil || !bytes.Contains(receiptScope, []byte(`"record_body_scope"`)) ||
		!bytes.Contains(receiptScope, []byte(`"presence":"optional"`)) {
		t.Fatalf("completeness receipt omitted the optional record contract: %s (%v)", receiptScope, err)
	}
	run, err := ExecuteComposition(context.Background(), "optional-records.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 12 || run.FiniteTotal != 12 || !run.RuntimeReplayed {
		t.Fatalf("optional record values did not survive execution: err=%v runtime=%+v", err, run)
	}
	graph, err := prepareCompositionGraph(context.Background(), "optional-records.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	for caseIndex, trace := range run.Traces {
		if len(trace.Deliveries) != 3 {
			t.Fatalf("case %d deliveries=%d", caseIndex, len(trace.Deliveries))
		}
		copy, relay, empty := trace.Deliveries[0], trace.Deliveries[1], trace.Deliveries[2]
		if !bytes.Equal(copy.Actual, relay.Input) || !bytes.Equal(copy.Actual, relay.Actual) ||
			!bytes.Equal(relay.Actual, empty.Input) || !bytes.Equal(empty.Actual, []byte(`{}`)) {
			t.Fatalf("case %d changed the optional record across the bind: %+v", caseIndex, trace.Deliveries)
		}
		for _, observation := range [][]CompositionRecordField{copy.InputFields, copy.ActualFields,
			relay.InputFields, relay.ActualFields, empty.InputFields, empty.ActualFields} {
			if len(observation) != 3 {
				t.Fatalf("case %d optional field observations=%d", caseIndex, len(observation))
			}
			for _, field := range observation {
				if field.Presence != "optional" || field.Present == nil || field.ID == "" {
					t.Fatalf("case %d lost optional field identity or presence: %+v", caseIndex, field)
				}
			}
		}
		for _, field := range empty.ActualFields {
			if *field.Present || field.Value != nil {
				t.Fatalf("case %d omitted field %q was materialized: %+v", caseIndex, field.Name, field)
			}
		}
		present := map[string]bool{"note": caseIndex == 1 || caseIndex == 2, "complete": caseIndex != 0, "count": caseIndex == 1 || caseIndex == 2}
		for _, field := range copy.ActualFields {
			if *field.Present != present[field.Name] {
				t.Fatalf("case %d field %q presence=%t", caseIndex, field.Name, *field.Present)
			}
			if !present[field.Name] && field.Value != nil {
				t.Fatalf("case %d absent field %q has a value: %s", caseIndex, field.Name, field.Value)
			}
		}
		if caseIndex == 1 {
			for _, field := range copy.ActualFields {
				if !*field.Present {
					t.Fatalf("explicit zero field %q became absent", field.Name)
				}
				want := map[string]string{"note": `""`, "complete": "false", "count": "0"}[field.Name]
				if string(field.Value) != want {
					t.Fatalf("explicit zero field %q = %s, want %s", field.Name, field.Value, want)
				}
			}
		}
	}
	for _, invalid := range []string{`{"note":null}`, `{"complete":0}`, `{"count":"0"}`, `{"unknown":"x"}`, `{"note":"x","note":"y"}`} {
		if _, err := graph.canonicalValue([]byte(invalid), "Profile"); err == nil {
			t.Fatalf("accepted invalid optional record: %s", invalid)
		}
	}
}
