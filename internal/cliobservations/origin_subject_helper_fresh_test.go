package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoooHelperSubject24131(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/return21917-runtime.json.gz"))
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
	saved := filepath.Join(t.TempDir(), "return24131.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "helper-reader")
	var out, diag bytes.Buffer
	if code := run([]string{"init", "--template", "origins", "--module", "example.org/team/helper24131", dir}, &out, &diag); code != exitOK {
		t.Fatal(code, diag.String())
	}
	guide, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil || !bytes.Contains(guide, []byte("--origin-root-activity-id")) {
		t.Fatal("embedded subject guide", err)
	}
	out.Reset()
	diag.Reset()
	args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-field-id", "urn:directreturn21917:tag", "--origin-activity-id", "urn:directreturn21917:make", "--origin-root-activity-id", "urn:directreturn21917:pick", "--origin-limit", "2", filepath.Join(dir, "gooo.workspace.json")}
	code := run(args, &out, &diag)
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		f, err := os.Create(filepath.Join(output, "helper-reader24131.json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		z := gzip.NewWriter(f)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if code != exitOK {
		t.Fatal(code, diag.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	p, r := receipt.OriginInput, receipt.Result.Runtime
	if p.SourceRows == nil || *p.SourceRows != 20 || p.TotalRows != 2 || p.ReturnedRows != 2 || p.Truncated || p.ProducerExecutions != 0 || p.ProducerModelCalls != 0 || len(r.Traces) != 2 || r.FiniteTotal != 0 || r.ModelCalls != 0 {
		t.Fatal("helper query scope", p)
	}
	for i, trace := range r.Traces {
		var actual struct {
			Status     string `json:"status"`
			FieldID    string `json:"field_id"`
			Activity   string `json:"original_activity_id"`
			Assignment int    `json:"assignment_index"`
			Filename   string `json:"filename"`
			Expression string `json:"expression"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("helper origin output")
		}
		status, expression := "FIELD_TARGET", "tag"
		if i == 1 {
			status = "ORIGINAL_EXPRESSION"
			expression = `"helper"`
		}
		if actual.Status != status || actual.Expression != expression || actual.Assignment != 3 || actual.FieldID != "urn:directreturn21917:tag" || actual.Activity != "urn:directreturn21917:make" || actual.Filename != "직접반환21917.gooo" {
			t.Fatal("original helper numbering and source", actual)
		}
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"help", "package"}, &out, &diag); code != exitOK || !strings.Contains(out.String(), "--origin-field-id") || !strings.Contains(out.String(), "source_rows") {
		t.Fatal("discoverable subject help", code, diag.String())
	}
	t.Log("real CLI root/actual-helper/field IDs conjoined;20 retained/2 selected/2 returned;original helper write3 retained;Gooo target/value classification and escaped-source expression;embedded/help guides;producer/model0;finite0")
}
