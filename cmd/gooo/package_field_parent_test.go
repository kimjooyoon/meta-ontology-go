package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func fieldParentFixture(t *testing.T) ([]byte, string) {
	t.Helper()
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "examples/local-field-bridge/recorded-runtime.json.gz"))
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
	return raw, root
}

func fieldParentInput(t *testing.T, raw []byte, shape string) (map[string]json.RawMessage, *bodyexecution.NativeOriginInput) {
	t.Helper()
	query := bodyexecution.NativeOriginQuery{Kind: "field-uses", Offset: 15, Limit: 1}
	encoded, report, err := packageOriginInputsWithShape(raw,
		packageruntime.EntrySpec{PackagePath: "origin/fieldparent", Activity: "Inspect"}, query, shape)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Inputs []map[string]map[string]json.RawMessage `json:"inputs"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil || len(document.Inputs) != 1 {
		t.Fatal("selected case input", err)
	}
	return document.Inputs[0]["origin/fieldparent:Inspect"], report
}

func TestFieldParentExecuteUsage(t *testing.T) {
	var output, diagnostic bytes.Buffer
	code := run([]string{"package", "execute"}, &output, &diagnostic)
	if code != exitUsage || !bytes.Contains(diagnostic.Bytes(), []byte("field-parent-columns")) {
		t.Fatalf("field-parent input shape missing from CLI usage: code=%d stderr=%s", code, diagnostic.String())
	}
}

func TestFieldParentProjectionIdentity(t *testing.T) {
	raw, _ := fieldParentFixture(t)
	for shape, count := range map[string]int{"field-lineage-columns": 26, "field-local-columns": 32, "field-parent-columns": 29} {
		row, report := fieldParentInput(t, raw, shape)
		if len(row) != count || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
			t.Fatal("projection compatibility", shape, len(row), report)
		}
	}
	first, report := fieldParentInput(t, raw, "field-parent-columns")
	if string(first["receipt_sha256"]) != `"`+workspaceDigest(raw)+`"` || report.ReceiptSHA256 != workspaceDigest(raw) {
		t.Fatal("raw-file identity was not transported")
	}
	spaced, _ := fieldParentInput(t, append(append([]byte(nil), raw...), '\n'), "field-parent-columns")
	if bytes.Equal(first["receipt_sha256"], spaced["receipt_sha256"]) || !bytes.Equal(first["journal_sha256"], spaced["journal_sha256"]) {
		t.Fatal("raw byte identity confused with field journal identity")
	}
	var runtime bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &runtime); err != nil {
		t.Fatal(err)
	}
	runtime.Traces[0].FieldWrites[14].ParentAssignment = 11
	changed, err := json.Marshal(runtime)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := fieldParentInput(t, changed, "field-parent-columns")
	if bytes.Equal(first["receipt_sha256"], other["receipt_sha256"]) || bytes.Equal(first["parents"], other["parents"]) {
		t.Fatal("distinct saved parent claims merged into one raw-file scope")
	}
	if !bytes.Equal(first["source_sha256"], other["source_sha256"]) || !bytes.Equal(first["journal_sha256"], other["journal_sha256"]) {
		t.Fatal("control must preserve the claimed source and journal identities")
	}
	if err := validatePackageOriginShape(bodyexecution.NativeOriginQuery{Kind: "field-uses", Limit: 2}, "field-parent-columns"); err == nil {
		t.Fatal("single-use projection silently accepted another page contract")
	}
}

func fieldParentExecute(t *testing.T, args []string) packageExecutionReceipt {
	t.Helper()
	var out, diagnostic bytes.Buffer
	code := run(args, &out, &diagnostic)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("Gooo field-parent execution", code, err, receipt.Error, diagnostic.String())
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		var compressed bytes.Buffer
		z := gzip.NewWriter(&compressed)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		name := "field-parent-controls.json.gz"
		if receipt.OriginInput != nil {
			name = "field-parent-saved-query.json.gz"
		}
		if err := os.WriteFile(filepath.Join(dir, name), compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return receipt
}

func TestFieldParentGoooControls(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	_, root := fieldParentFixture(t)
	receipt := fieldParentExecute(t, []string{"package", "execute", "--json", "--cases",
		filepath.Join(root, "examples/field-parent-origins/controls.json"),
		filepath.Join(root, "examples/field-parent-origins.workspace.json")})
	r := receipt.Result.Runtime
	if receipt.Decision != "PASS" || r.FiniteTotal != 21 || r.FinitePassed != 21 || r.ModelCalls != 0 || !r.RuntimeReplayed {
		t.Fatal("authored full-output control population", r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}

func TestFieldParentGoooSavedQuery(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	raw, root := fieldParentFixture(t)
	saved := filepath.Join(t.TempDir(), "recorded.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	receipt := fieldParentExecute(t, []string{"package", "execute", "--json", "--origin-receipt", saved,
		"--origin-shape", "field-parent-columns", "--origin-offset", "15",
		filepath.Join(root, "examples/field-parent-origins.workspace.json")})
	r := receipt.Result.Runtime
	if receipt.Decision != "OBSERVED" || r.FiniteTotal != 0 || r.ModelCalls != 0 || receipt.OriginInput == nil || receipt.OriginInput.ProducerExecutions != 0 || len(r.Traces) != 1 || len(r.Traces[0].Deliveries) != 1 {
		t.Fatal("saved query scope")
	}
	var result struct {
		Status string   `json:"status"`
		Child  int      `json:"child"`
		Parent int      `json:"parent"`
		Lines  []string `json:"lines"`
	}
	if err := json.Unmarshal(r.Traces[0].Deliveries[0].Actual, &result); err != nil || result.Status != "RECORDED_DIRECT_PARENT" || result.Child != 14 || result.Parent != 17 || len(result.Lines) != 4 {
		t.Fatal("literal saved field-copy expectation", result, err)
	}
	base := "<urn:gooo:field-parent:v1:saved:" + workspaceDigest(raw) + ":case:0:field:write:"
	expected := base + "14> <urn:gooo:field-parent:v1:retainedParent> " + base + "17> ."
	if result.Lines[2] != expected {
		t.Fatal("raw file/case/field/write identity", result.Lines[2])
	}
}
