package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordAssemblyWorkerOwnsConcurrentSelectionAndCheckpoint(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var input, output bytes.Buffer
	for range 4 {
		raw, _ := json.Marshal(Request{Schema: RequestSchema, CorrelationID: "record", Source: string(source), Activity: "Select"})
		input.Write(raw)
		input.WriteByte('\n')
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	if err = Run(context.Background(), g, io.NopCloser(&input), asWriteCloser(&output), 2); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, output.Bytes())
	if len(results) != 4 {
		t.Fatal("record worker responses", len(results))
	}
	for _, result := range results {
		if result.Status != "completed" || result.Response == nil || result.Response.Report.RecordAssembly.FieldsPassed != 15 || result.Response.Report.RecordAssembly.ModelCalls != 0 {
			t.Fatal("record worker lost finite state", result.Error)
		}
		if _, err = bodycodegen.RealizeSourceAssembly(context.Background(), "stream.gooo", source, *result.Response); err != nil {
			t.Fatal(err)
		}
	}
}
