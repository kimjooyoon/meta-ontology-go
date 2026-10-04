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
