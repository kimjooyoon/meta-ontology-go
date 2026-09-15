package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

type profileReader struct{ source string }

func (reader profileReader) ReadFile(string) ([]byte, error) { return []byte(reader.source), nil }

type profileMeasurer struct{ sequence int }

func (value *profileMeasurer) Measure(run func() sourceexecution.Receipt) (sourceexecution.Receipt, languageprofile.Measurement) {
	value.sequence++
	return run(), languageprofile.Measurement{WallNanoseconds: int64(100 + value.sequence), TotalAllocBytes: uint64(1000 + value.sequence)}
}

func TestRunProfileProducesStructuredReceipt(t *testing.T) {
	source := "package billing\nnamespace billing\nentity Order id \"billing://order\"\nentity Receipt id \"billing://receipt\"\nactivity PayOrder(Order) -> Receipt\n"
	var stdout, stderr bytes.Buffer
	code := runProfile([]string{"--json", "--samples", "3", "--entry", "PayOrder", "billing.gooo"},
		profileReader{source}, &profileMeasurer{}, &stdout, &stderr)
	var receipt languageprofile.Receipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || stderr.Len() != 0 || receipt.Decision != "PASS" || receipt.Summary.SamplesObserved != 3 {
		t.Fatalf("code=%d stderr=%q receipt=%#v", code, stderr.String(), receipt)
	}
}
