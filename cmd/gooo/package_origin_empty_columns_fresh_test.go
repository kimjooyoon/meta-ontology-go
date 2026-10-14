package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestEmptyAssemblyColumns12229(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	files := map[string]string{
		"main.gooo": `package emptycols
namespace emptycols12229
entity Integer id "urn:gooo:type:integer"
activity Build(Integer) -> Integer id "urn:emptycols12229:build" computes "let adjusted=input+661;return adjusted"
`,
		"w.json":     `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/12229","activity":"Build"},"packages":[{"path":"origin/12229","name":"emptycols","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/12229:Build":673},"expected":{"origin/12229:Build":1334}},{"inputs":{"origin/12229:Build":-677},"expected":{"origin/12229:Build":-16}}]}`}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var producer, consumer, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "w.json")}, &producer, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), producer.String())
	}
	saveColumnsOutput12161(t, "empty12229-producer.json", producer.Bytes())
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
		t.Fatal("empty assembly table", code, diagnostics.String(), consumer.String())
	}
	checkEmptyColumns12229(t, producer.Bytes(), consumer.Bytes())
	saveColumnsOutput12161(t, "empty12229-consumer.json", consumer.Bytes())
	entry := packageruntime.EntrySpec{PackagePath: "origincoverage", Activity: "Summarize"}
	query := bodyexecution.NativeOriginQuery{Kind: "local-fill-fragments", Limit: 32}
	if _, _, err := packageOriginInputs(producer.Bytes(), entry, query); err == nil {
		t.Fatal("row-input API silently changed its empty-table contract")
	}
	query.Offset = 1
	if _, _, err := packageOriginInputsWithShape(producer.Bytes(), entry, query, "local-fill-columns"); err == nil {
		t.Fatal("an invalid offset was reported as an empty table")
	}
	t.Log("new direct-authored producer2/2;zero assembly fragments/two recorded writes;Gooo assembly-origin coverage0%;one consumer row;no original producer/model calls")
}

func checkEmptyColumns12229(t *testing.T, producer, consumer []byte) {
	t.Helper()
	var p, c packageExecutionReceipt
	if err := json.Unmarshal(producer, &p); err != nil || p.Result == nil || p.Result.Runtime.FinitePassed != 2 {
		t.Fatal("fresh producer", err)
	}
	if err := json.Unmarshal(consumer, &c); err != nil || c.Result == nil || c.OriginInput == nil {
		t.Fatal(err)
	}
	input := c.OriginInput
	if input.ReturnedRows != 0 || input.TotalRows != 0 || input.ConsumerRows != 1 || input.Truncated || input.ProducerExecutions != 0 || input.ProducerModelCalls != 0 {
		t.Fatal("empty source table lost its explicit population", input)
	}
	if len(c.Result.Runtime.Traces) != 1 || len(c.Result.Runtime.Traces[0].Deliveries) != 1 || c.Result.Runtime.ModelCalls != 0 || c.Result.Runtime.FiniteTotal != 0 {
		t.Fatal("new Gooo consumer observation scope")
	}
	var actual struct {
		Writes  int    `json:"recorded_writes"`
		Mapped  int    `json:"mapped_writes"`
		Missing int    `json:"missing_writes"`
		Ratio   int    `json:"coverage_basis_points"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(c.Result.Runtime.Traces[0].Deliveries[0].Actual, &actual); err != nil || actual.Writes != 2 || actual.Mapped != 0 || actual.Missing != 2 || actual.Ratio != 0 || actual.Status != "complete-recorded" {
		t.Fatal("zero assembly coverage is not an unknown ratio", err, actual)
	}
}
