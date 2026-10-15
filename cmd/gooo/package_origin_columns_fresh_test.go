package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSavedOriginColumns12161(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	producer := produceColumnSource12161(t)
	receipt := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(receipt, producer, 0600); err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	manifest := filepath.Join(root, "examples/derived-origin/coverage.workspace.json")
	for _, page := range []struct {
		name, offset, limit, status                string
		rows, mapped, candidates, templates, ratio int
	}{{"full", "0", "4", "complete-recorded", 4, 4, 2, 2, 10000},
		{"middle", "1", "2", "partial-page", 2, 2, 1, 1, -1},
		{"tail", "3", "1", "partial-page", 1, 1, 0, 1, -1}} {
		var output, diagnostics bytes.Buffer
		args := []string{"package", "execute", "--json", "--origin-receipt", receipt,
			"--origin-shape", "local-fill-columns", "--origin-offset", page.offset, "--origin-limit", page.limit, manifest}
		if code := run(args, &output, &diagnostics); code != exitOK {
			t.Fatal(page.name, code, diagnostics.String(), output.String())
		}
		checkColumnsResult12161(t, output.Bytes(), page.rows, page.mapped, page.candidates, page.templates, page.ratio, page.status)
		saveColumnsOutput12161(t, "columns12161-"+page.name+".json", output.Bytes())
	}
	t.Log("fresh producer2/2; direct saved receipt->columns->Gooo coverage full/middle/tail3/3; fact rows4/2/1 and consumer rows1 distinguished; original producer executions during reads0/models0")
}

func produceColumnSource12161(t *testing.T) []byte {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"main.gooo": `package origincols
namespace origincols12161
entity Integer id "urn:gooo:type:integer"
activity Build(Integer) -> Integer id "urn:origincols12161:build" computes "let chosen=__GOOO_BODY_HOLE_value__;let copied=chosen+3;return copied" assembling {
 source_fill intent "소스의 두 대입을 저장하고 Gooo가 출처 연결 비율을 계산한다" {
  hole "value"
  candidate "plus12161" {fill "value" "input+631"}
  candidate "minus12161" {fill "value" "input-641"}
 }
 case "643" -> "1277"
 case "-647" -> "-13"
}
`,
		"gooo.workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/12161","activity":"Build"},"packages":[{"path":"origin/12161","name":"origincols","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json":          `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/12161:Build":653},"expected":{"origin/12161:Build":1287}},{"inputs":{"origin/12161:Build":-659},"expected":{"origin/12161:Build":-25}}]}`}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal("fresh producer", code, diagnostics.String(), output.String())
	}
	saveColumnsOutput12161(t, "columns12161-producer.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 {
		t.Fatal("producer denominator", err)
	}
	return output.Bytes()
}

func checkColumnsResult12161(t *testing.T, raw []byte, rows, mapped, candidates, templates, ratio int, status string) {
	t.Helper()
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err)
	}
	input := receipt.OriginInput
	if input.Query.Kind != "local-fill-fragments" || input.TotalRows != 4 || input.ReturnedRows != rows || input.ConsumerRows != 1 || input.InputShape != "gooo/local-fill-columns/v1" || input.ProducerExecutions != 0 || input.ProducerModelCalls != 0 {
		t.Fatal("source page and consumer envelope populations", input)
	}
	runtime := receipt.Result.Runtime
	if len(runtime.Traces) != 1 || len(runtime.Traces[0].Deliveries) != 1 || runtime.ModelCalls != 0 || runtime.FiniteTotal != 0 {
		t.Fatal("new consumer observation scope")
	}
	var result struct {
		Mapped     int    `json:"mapped_writes"`
		Candidates int    `json:"candidate_fragments"`
		Templates  int    `json:"template_fragments"`
		Ratio      int    `json:"coverage_basis_points"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(runtime.Traces[0].Deliveries[0].Actual, &result); err != nil || result.Mapped != mapped || result.Candidates != candidates || result.Templates != templates || result.Ratio != ratio || result.Status != status {
		t.Fatal("Gooo metric output", err, result)
	}
}

func TestOriginColumnOptions12187(t *testing.T) {
	for _, args := range [][]string{
		{"--inputs", "input.json", "--origin-shape", "local-fill-columns", "w.json"},
		{"--origin-receipt", "r.json", "--origin-shape", "unknown", "w.json"},
		{"--origin-receipt", "r.json", "--origin-kind", "reads", "--origin-shape", "local-fill-columns", "w.json"},
	} {
		if _, _, err := parsePackageExecuteArgs(args); err == nil {
			t.Fatal("incompatible source-column request accepted", args)
		}
	}
	flags, _, err := parsePackageExecuteArgs([]string{"--origin-receipt", "r.json", "--origin-shape", "local-fill-columns", "w.json"})
	if err != nil {
		t.Fatal(err)
	}
	query, err := packageOriginQuery(flags)
	if err != nil || query.Kind != "local-fill-fragments" || query.Limit != 32 {
		t.Fatal("shape did not retain its explicit fragment profile", err, query)
	}
	t.Log("three incompatible shape requests rejected; explicit column profile selects fragment kind")
}

func saveColumnsOutput12161(t *testing.T, name string, raw []byte) {
	t.Helper()
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
