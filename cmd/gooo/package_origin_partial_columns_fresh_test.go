package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPartialOriginColumns12317(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	files := map[string]string{
		"main.gooo": `package partialcols
namespace partialcols12317
entity Integer id "urn:gooo:type:integer"
activity Advance(Integer) -> Integer id "urn:partialcols12317:advance" computes "let corrected=input+683;let doubled=corrected+corrected;return doubled"
activity Build(Integer) -> Integer id "urn:partialcols12317:build" computes "let count=0;let sum=input;for count=0;count<1024;count=count+1 {let chunk=Advance(count);sum=sum+chunk};return sum"
`,
		"w.json":     `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/12317","activity":"Build"},"packages":[{"path":"origin/12317","name":"partialcols","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/12317:Build":691},"expected":{"origin/12317:Build":2447027}}]}`,
	}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var producer, consumer, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "w.json")}, &producer, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), producer.String())
	}
	saveColumnsOutput12161(t, "partial12317-producer.json", producer.Bytes())
	var p packageExecutionReceipt
	if err := json.Unmarshal(producer.Bytes(), &p); err != nil || p.Result == nil || p.Result.Runtime.FinitePassed != 1 {
		t.Fatal("new helper-loop producer", err)
	}
	summary := p.Result.Runtime.LocalObservation
	if summary == nil || summary.Status != "PARTIAL" || summary.RecordedWrites != 4096 || summary.WritesDropped == 0 {
		t.Fatal("actual recording did not exceed its bound", summary)
	}
	receipt := filepath.Join(root, "saved.json")
	if err := os.WriteFile(receipt, producer.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	manifest := filepath.Join(sourceRoot, "examples/derived-origin/coverage.workspace.json")
	if code := run([]string{"package", "execute", "--json", "--origin-receipt", receipt, "--origin-shape", "local-fill-columns", manifest}, &consumer, &diagnostics); code != exitOK {
		t.Fatal("partial recorded table", code, diagnostics.String(), consumer.String())
	}
	saveColumnsOutput12161(t, "partial12317-consumer.json", consumer.Bytes())
	checkPartialColumns12317(t, consumer.Bytes())
	t.Logf("new helper-loop producer1/1;4096 writes retained/%d dropped;Gooo partial-recording keeps ratio and missing population unspecified;original producer/model calls during read0", summary.WritesDropped)
}

func checkPartialColumns12317(t *testing.T, raw []byte) {
	t.Helper()
	var c packageExecutionReceipt
	if err := json.Unmarshal(raw, &c); err != nil || c.Result == nil || c.OriginInput == nil {
		t.Fatal(err)
	}
	input := c.OriginInput
	if input.LocalObservation == nil || input.LocalObservation.Status != "PARTIAL" || input.ReturnedRows != 0 || input.ConsumerRows != 1 || input.Truncated || input.ProducerExecutions != 0 || input.ProducerModelCalls != 0 {
		t.Fatal("recording partiality was lost", input)
	}
	runtime := c.Result.Runtime
	if len(runtime.Traces) != 1 || len(runtime.Traces[0].Deliveries) != 1 || runtime.ModelCalls != 0 || runtime.FiniteTotal != 0 {
		t.Fatal("consumer observation scope")
	}
	var result struct {
		Writes  int    `json:"recorded_writes"`
		Missing int    `json:"missing_writes"`
		Ratio   int    `json:"coverage_basis_points"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(runtime.Traces[0].Deliveries[0].Actual, &result); err != nil || result.Writes != 4096 || result.Missing != -1 || result.Ratio != -1 || result.Status != "partial-recording" {
		t.Fatal("partial record became a complete population", err, result)
	}
}
