package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageLocalFillReader11553(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	files := map[string]string{"main.gooo": localFillSource11553,
		"gooo.workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/11553","activity":"Sum"},"packages":[{"path":"origin/11553","name":"origin","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json":          `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/11553:Sum":467},"expected":{"origin/11553:Sum":916}},{"inputs":{"origin/11553:Sum":-479},"expected":{"origin/11553:Sum":928}}]}`}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var producer, consumer, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &producer, &diagnostics); code != exitOK {
		t.Fatal("new producer", code, diagnostics.String(), producer.String())
	}
	receipt := filepath.Join(root, "saved.json")
	if err := os.WriteFile(receipt, producer.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "producer11553.json"), producer.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	reader := filepath.Join(sourceRoot, "examples/native-origin-reader/local-fill/reader.workspace.json")
	if code := run([]string{"package", "execute", "--json", "--origin-receipt", receipt, "--origin-kind", "local-fill-fragments", "--origin-limit", "2", reader}, &consumer, &diagnostics); code != exitOK {
		t.Fatal("saved fragment reader", code, diagnostics.String(), consumer.String())
	}
	checkLocalFillCLI11553(t, producer.Bytes(), consumer.Bytes())
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		for name, raw := range map[string][]byte{"producer11553.json": producer.Bytes(), "consumer11553.json": consumer.Bytes()} {
			if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("fresh CLI producer2/2; saved-origin CLI two classifications match candidate/template; producer executions during read0; models0")
}

func checkLocalFillCLI11553(t *testing.T, producer, consumer []byte) {
	t.Helper()
	var p, c packageExecutionReceipt
	if err := json.Unmarshal(producer, &p); err != nil || p.Result == nil || p.Result.Runtime.FinitePassed != 2 {
		t.Fatal("producer finite results", err)
	}
	if err := json.Unmarshal(consumer, &c); err != nil || c.Result == nil || c.OriginInput == nil {
		t.Fatal("consumer receipt", err)
	}
	if c.OriginInput.TotalRows != 6 || c.OriginInput.ReturnedRows != 2 || c.OriginInput.ProducerExecutions != 0 || c.OriginInput.ProducerModelCalls != 0 || c.Result.Runtime.ModelCalls != 0 {
		t.Fatal("recorded input population", c.OriginInput)
	}
	for index, want := range []string{"candidate", "template"} {
		trace := c.Result.Runtime.Traces[index]
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want {
			t.Fatal("Gooo classification", index, trace)
		}
	}
}

const localFillSource11553 = `package origin
namespace receipt11553
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer) -> Integer id "urn:receipt11553:sum" computes "let mark=__GOOO_BODY_HOLE_mark__;let kept=mark;if input>0 {mark=kept+input}else{mark=kept-input};return mark" assembling {
 source_fill intent "저장된 대입에서 실제로 선택한 표현의 원래 출처를 읽는다" {
  hole "mark"
  candidate "fixed11553" {fill "mark" "449"}
  candidate "other11553" {fill "mark" "457"}
 }
 case "461" -> "910"
 case "-463" -> "912"
}
`
