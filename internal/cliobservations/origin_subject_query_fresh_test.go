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

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooOriginSubjects23891(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/choice22133-runtime.json.gz"))
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
	var original bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "saved23891.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "reader")
	var out, diagnostic bytes.Buffer
	if code := run([]string{"init", "--template", "origins", "--module", "example.org/team/subject23891", dir}, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String())
	}
	workspace := filepath.Join(dir, "gooo.workspace.json")
	out.Reset()
	diagnostic.Reset()
	field := "urn:assembledlength22133:total"
	args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-field-id", field, "--origin-offset", "1", "--origin-limit", "3", workspace}
	code := run(args, &out, &diagnostic)
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		f, err := os.Create(filepath.Join(output, "subject-reader23891.json.gz"))
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
		t.Fatal(code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	r, p := receipt.Result.Runtime, receipt.OriginInput
	if receipt.Decision != "OBSERVED" || p.Query.Kind != "field-input-locations" || p.Query.FieldID != field || p.SourceRows == nil || *p.SourceRows != 27 || p.TotalRows != 9 || p.ReturnedRows != 3 || p.NextOffset == nil || *p.NextOffset != 4 || !p.Truncated || p.ProducerExecutions != 0 || p.ProducerModelCalls != 0 || r.FiniteTotal != 0 || r.ModelCalls != 0 || len(r.Traces) != 3 {
		t.Fatal("exact ID selection,matched page and producer scope", p)
	}
	binding := original.FieldFileSources.Bindings[0]
	locations := []bodyexecution.CompositionLocalFileLocation{binding.Value, binding.Choice.Value, binding.Target}
	statuses := []string{"UNRESOLVED", "DECLARED_ALTERNATIVE", "FIELD_TARGET"}
	for i, trace := range r.Traces {
		var actual struct {
			Status     string `json:"status"`
			Field      string `json:"field_id"`
			Filename   string `json:"filename"`
			Start      int    `json:"source_start"`
			End        int    `json:"source_end"`
			Expression string `json:"expression"`
			Assignment int    `json:"assignment_index"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("Gooo subject output")
		}
		want := locations[i]
		if actual.Status != statuses[i] || actual.Field != field || actual.Filename != want.Filename || actual.Start != want.Start || actual.End != want.End || actual.Expression != want.Expression || actual.Assignment != 0 {
			t.Fatal("original locations after selection", i, actual, want)
		}
	}
	for _, flag := range []string{"--origin-field-id", "--origin-activity-id", "--origin-root-activity-id"} {
		out.Reset()
		diagnostic.Reset()
		if code := run([]string{"package", "execute", "--inputs", "missing.json", flag, "urn:some:id", workspace}, &out, &diagnostic); code == exitOK || !strings.Contains(diagnostic.String(), "origin query options require --origin-receipt") {
			t.Fatal("subject options without a receipt", flag, code, diagnostic.String())
		}
	}
	out.Reset()
	diagnostic.Reset()
	if code := run([]string{"package", "execute", "--origin-receipt", saved, "--origin-kind", "reads", "--origin-field-id", field, workspace}, &out, &diagnostic); code == exitOK || !strings.Contains(diagnostic.String(), "origin subject IDs require") {
		t.Fatal("incompatible subject kind", code, diagnostic.String())
	}
	t.Log("real CLI generated a distinct Gooo reader;exact field default kind;27 retained/9 selected/3 paged;original source coordinates and write numbers;three no-receipt filters and wrong kind rejected;producer/model0;input-only finite0")
}
