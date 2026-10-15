package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageLocalFillReader11591(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	files := map[string]string{"main.gooo": localFillSource11591,
		"gooo.workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/11591","activity":"Sum"},"packages":[{"path":"origin/11591","name":"origin","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json":          `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/11591:Sum":509},"expected":{"origin/11591:Sum":465}},{"inputs":{"origin/11591:Sum":-521},"expected":{"origin/11591:Sum":453}}]}`}
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
		if err := os.WriteFile(filepath.Join(dir, "producer11591.json"), producer.Bytes(), 0600); err != nil {
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
		for name, raw := range map[string][]byte{"producer11591.json": producer.Bytes(), "consumer11591.json": consumer.Bytes()} {
			if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("fresh CLI producer2/2; saved-origin CLI two classifications match candidate/template; producer executions during read0; models0")
}

const localFillSource11591 = `package origin
namespace receipt11591
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer) -> Integer id "urn:receipt11591:sum" computes "let mark=__GOOO_BODY_HOLE_mark__;let kept=mark*2;if input>0 {mark=kept-input}else{mark=kept+input};return mark" assembling {
 source_fill intent "저장된 대입에서 실제로 선택한 표현의 원래 출처를 읽는다" {
  hole "mark"
  candidate "fixed11591" {fill "mark" "487"}
  candidate "other11591" {fill "mark" "491"}
 }
 case "499" -> "475"
 case "-503" -> "471"
}
`
