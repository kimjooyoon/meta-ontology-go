package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageOriginsReadOnlyCLI5663(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	dir := t.TempDir()
	source := `package credit
namespace credit5663
entity Integer id "urn:gooo:type:integer"
activity Credit(Integer) -> Integer id "urn:credit5663:credit" computes "let credited = __GOOO_BODY_HOLE_credit__; return credited" assembling {
 source_fill intent "선택한 가산 규칙이 반환값으로 이어진다" {
  hole "credit"
  candidate "credit5663" { fill "credit" "input + 59" }
  candidate "other5663" { fill "credit" "input" }
 }
 case "31" -> "90"
}
`
	files := map[string]string{"credit.gooo.fixture": source, "workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"credit5663","activity":"Credit"},"packages":[{"path":"credit5663","name":"credit","sources":["credit.gooo.fixture"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"credit5663:Credit":37},"expected":{"credit5663:Credit":96}},{"inputs":{"credit5663:Credit":43},"expected":{"credit5663:Credit":102}}]}`}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(dir, "cases.json"), filepath.Join(dir, "workspace.json")}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	raw := append([]byte(nil), out.Bytes()...)
	path := filepath.Join(dir, "saved.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOOO_LAYA_URL", "http://127.0.0.1:1/query-must-not-call")
	t.Setenv("GOOO_LAYA_API_KEY", "unused")
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"package", "origins", "--receipt", path, "--candidate", "credit5663", "--json"}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var query packageOriginsReceipt
	if err := json.Unmarshal(out.Bytes(), &query); err != nil {
		t.Fatal(err)
	}
	if query.ReceiptSHA256 != workspaceDigest(raw) || len(query.Report.Rows) != 1 || query.Report.Rows[0].Use.RootKind != "return" {
		t.Fatal("query receipt", query)
	}
	if query.Report.Rows[0].Locations[0].View != "assembly_input" {
		t.Fatal("invented original file for return-only record")
	}
	jsonOutput := append([]byte(nil), out.Bytes()...)
	out.Reset()
	if code := run([]string{"package", "origins", "--receipt", path}, &out, &diagnostics); code != exitOK || !strings.Contains(out.String(), "input + 59") || !strings.Contains(out.String(), "credited") {
		t.Fatal("text origin path", code, diagnostics.String(), out.String())
	}
	textOutput := out.String()
	out.Reset()
	if code := run([]string{"help", "package", "origins"}, &out, &diagnostics); code != exitOK || !strings.Contains(out.String(), packageOriginsUsage) {
		t.Fatal("origin help", code, out.String())
	}
	retained, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, retained) {
		t.Fatal("query rewrote saved result", err)
	}
	if output := os.Getenv("GOOO_ORIGIN_QUERY_OUTPUT"); output != "" {
		for name, content := range map[string][]byte{"execution5663.json": raw, "query5663.json": jsonOutput, "query5663.txt": []byte(textOutput)} {
			if err := os.WriteFile(filepath.Join(output, name), content, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("5663: fresh native2/2; JSON and text origin lookup preserve receipt; return-only origin stays assembly-input scoped; environment provider unused; dedicated help available")
}

func TestPackageOriginsArgumentBounds5689(t *testing.T) {
	for _, args := range [][]string{{}, {"--receipt", "x", "--limit", "0"}, {"--receipt", "x", "--limit", "129"}, {"--receipt", "x", "--candidate"}, {"--receipt", "x", "--receipt", "y"}} {
		if _, _, err := parsePackageOriginArgs(args); err == nil {
			t.Fatal("invalid arguments accepted", args)
		}
	}
	path, query, err := parsePackageOriginArgs([]string{"--candidate", "selected5689", "--receipt", "record.json", "--activity", "urn:activity5689", "--limit", "7"})
	if err != nil || path != "record.json" || query.Limit != 7 || query.ActivityID != "urn:activity5689" || query.CandidateID != "selected5689" {
		t.Fatal(path, query, err)
	}
}
