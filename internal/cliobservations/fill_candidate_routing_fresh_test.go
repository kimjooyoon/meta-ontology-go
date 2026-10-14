package cliobservations

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCandidateOriginRouting24871(t *testing.T) {
	file := filepath.Join(t.TempDir(), "普通24871.gooo")
	source := []byte(`package ordinary24871
namespace ordinary24871
entity Integer id "urn:gooo:type:integer"
activity Shift(Integer)->Integer id "urn:ordinary24871:shift" computes "return input+24871"
`)
	if err := os.WriteFile(file, source, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := run([]string{"body-origins", "--activity", "Shift", file}, &out, &diag); code != exitOK {
		t.Fatal("ordinary typed origin route", code, diag.String(), out.String())
	}
	var report struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(out.Bytes(), &report) != nil || report.Status != "RESOLVED" {
		t.Fatal("typed origin report", out.String())
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"body-origins", "--activity", "Shift", "--fill-candidate", "candidate24871", file}, &out, &diag); code == exitOK || !strings.Contains(diag.String(), "source fill declaration") {
		t.Fatal("ordinary body gained a candidate declaration", code, diag.String())
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"body-origins", "--activity", "Shift", "--fill-candidate", "first24871", "--fill-candidate", "second24871", file}, &out, &diag); code == exitOK || !strings.Contains(diag.String(), "usage:") {
		t.Fatal("duplicate candidate selector", code, diag.String())
	}
	t.Log("routing24871:new ordinary source retains typed-origin route;absent source_fill and duplicate selector rejected;model0")
}
