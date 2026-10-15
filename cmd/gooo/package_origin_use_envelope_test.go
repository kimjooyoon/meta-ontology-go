package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestRecordedUsageOuterEnvelopeBinding(t *testing.T) {
	f, err := os.Open("../../examples/recorded-use-origins/full-producer-receipt.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	query := bodyexecution.NativeOriginQuery{Kind: "local-uses", Offset: 14, Limit: 1}
	encoded, report, err := packageOriginInputsWithShape(raw, packageruntime.EntrySpec{PackagePath: "origin/recordeduse", Activity: "Inspect"}, query, "local-use-columns")
	if err != nil {
		t.Fatal(err)
	}
	var inputs struct {
		Inputs []map[string]bodyexecution.LocalUseColumns `json:"inputs"`
	}
	if err := json.Unmarshal(encoded, &inputs); err != nil || len(inputs.Inputs) != 1 {
		t.Fatal("column envelope", err)
	}
	columns := inputs.Inputs[0]["origin/recordeduse:Inspect"]
	if columns.ReceiptSHA256 != workspaceDigest(raw) || report.ReceiptSHA256 != columns.ReceiptSHA256 || columns.Cases[0] != 1 || columns.UseIndices[0] != 6 || columns.Assignments[0] != 0 {
		t.Fatal("full saved JSON scope changed", columns.ReceiptSHA256, report.ReceiptSHA256)
	}
	if report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("metadata binding executed its producer")
	}
}
