package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedExplanationControlledCases(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/recorded-use-origins/explanation-cases.json",
		"../../examples/recorded-use-origins/explanation.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("Gooo explanation", code, err, receipt.Error, diag.String())
	}
	if r := receipt.Result.Runtime; receipt.Decision != "PASS" || r.FiniteTotal != 10 || r.FinitePassed != 10 || r.ModelCalls != 0 {
		t.Fatal("controlled explanation outcomes", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}

func TestRecordedExplanationSavedRead(t *testing.T) {
	file := recordedExplanationReceipt(t)
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", file, "--origin-shape", "local-use-columns",
		"--origin-offset", "0", "../../examples/recorded-use-origins/explanation.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("saved explanation", code, err, receipt.Error, diag.String())
	}
	r, origin := receipt.Result.Runtime, receipt.OriginInput
	if receipt.Decision != "OBSERVED" || origin.ReceiptSHA256 != "sha256:a15f33ca09f99265aa9b838dcd4628a9dd74a8b77dcff93e1156241738a5441a" ||
		origin.Query.Offset != 0 || origin.Query.Limit != 1 || origin.TotalRows != 32 || origin.ReturnedRows != 1 ||
		origin.ProducerExecutions != 0 || origin.ProducerModelCalls != 0 || r.ModelCalls != 0 {
		t.Fatal("explanation lost saved scope or executed producer/model", receipt.Decision, origin)
	}
	if len(r.Traces) != 1 || len(r.Traces[0].Deliveries) != 1 {
		t.Fatal("one selected entry result required")
	}
	var text string
	if err := json.Unmarshal(r.Traces[0].Deliveries[0].Actual, &text); err != nil ||
		text != "호출이 'value'의 기록된 값 버전을 읽었습니다. 같은 기록에서 이 버전의 대입 완료를 확인했습니다." {
		t.Fatal("saved read explanation differs", text, err)
	}
}

func recordedExplanationReceipt(t *testing.T) string {
	t.Helper()
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
	file := filepath.Join(t.TempDir(), "saved-producer.json")
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}
