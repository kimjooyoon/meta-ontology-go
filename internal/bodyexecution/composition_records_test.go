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
