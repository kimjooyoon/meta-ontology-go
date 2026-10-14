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

func TestOriginsStarter23117(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	for _, module := range []string{"", "example.org/team/roots23117"} {
		dir := filepath.Join(t.TempDir(), "reader")
		args := []string{"init", "--template", "origins"}
		if module != "" {
			args = append(args, "--module", module)
		}
		var out, diagnostic bytes.Buffer
		if code := run(append(args, dir), &out, &diagnostic); code != exitOK {
			t.Fatal("new reader", code, out.String(), diagnostic.String())
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 6 {
			t.Fatal("six self-contained starter files", len(entries), err)
		}
		for _, entry := range entries {
			raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil || strings.Contains(string(raw), "{{") {
				t.Fatal("unexpanded file", entry.Name(), err)
			}
		}
		source := filepath.Join(dir, "inspect.gooo")
		before, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		prefix := module
		if prefix == "" {
			prefix = "origins"
		}
		if !bytes.Contains(before, []byte("gooo://"+prefix+"/origins/inspect")) {
			t.Fatal("module-bound semantic ID")
		}
		out.Reset()
		diagnostic.Reset()
		if code := run(append(args, dir), &out, &diagnostic); code == exitOK {
			t.Fatal("existing project accepted")
		}
		after, err := os.ReadFile(source)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("existing source changed", err)
		}
		out.Reset()
		diagnostic.Reset()
		if code := run([]string{"package", "interface", "--json", filepath.Join(dir, "gooo.workspace.json")}, &out, &diagnostic); code != exitOK || out.Len() == 0 {
			t.Fatal("inspect generated workspace", code, diagnostic.String())
		}
		if module != "" {
			observeOriginsStarter23117(t, dir)
		}
	}
	for _, module := range []string{"../escape", "Upper/Name", "example.org//missing"} {
		dir := filepath.Join(t.TempDir(), "invalid")
		var out, diagnostic bytes.Buffer
		code := run([]string{"init", "--template", "origins", "--module", module, dir}, &out, &diagnostic)
		if code == exitOK || !strings.Contains(diagnostic.String(), "invalid origins module path") {
			t.Fatal("invalid module accepted", module, code, diagnostic.String())
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatal("invalid module wrote destination", err)
		}
	}
	t.Log("default/custom modules create six local files with stable IDs;existing projects preserved;three invalid modules rejected before writing;model0")
}

func observeOriginsStarter23117(t *testing.T, dir string) {
	t.Helper()
	var out, diagnostic bytes.Buffer
	workspace := filepath.Join(dir, "gooo.workspace.json")
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(dir, "cases.json"), workspace}, &out, &diagnostic); code != exitOK {
		t.Fatal("practice cases", code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, receipt.Error)
	}
	if receipt.Decision != "PASS" || receipt.Result.Runtime.FinitePassed != 4 || receipt.Result.Runtime.FiniteTotal != 4 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("fictional practice cases4/4", receipt.Result.Runtime.FinitePassed)
	}
	saveOriginsStarter23117(t, "practice23117.json.gz", out.Bytes())
	path := filepath.Join(dir, "execution.json")
	if err := os.WriteFile(path, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	moved := dir + "-moved"
	if err := os.Rename(dir, moved); err != nil {
		t.Fatal(err)
	}
	dir = moved
	workspace = filepath.Join(dir, "gooo.workspace.json")
	out.Reset()
	diagnostic.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", filepath.Join(dir, "execution.json"), "--inputs", filepath.Join(dir, "inputs.json"), workspace}, &out, &diagnostic); code != exitOK {
		t.Fatal("moved input-only replay", code, diagnostic.String(), out.String())
	}
	receipt = packageExecutionReceipt{}
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Replay == nil || receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || len(receipt.Result.Runtime.Traces) != 1 || !bytes.Contains(out.Bytes(), []byte("EDIT_THIS_SAMPLE")) {
		t.Fatal("portable input-only replay", err, receipt.Error)
	}
	saveOriginsStarter23117(t, "moved-input23117.json.gz", out.Bytes())
	observeActualStarterOrigins23117(t, workspace)
}

func observeActualStarterOrigins23117(t *testing.T, workspace string) {
	t.Helper()
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
	path := filepath.Join(t.TempDir(), "original.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "field-input-locations", "--origin-limit", "9", workspace}, &out, &diagnostic)
	saveOriginsStarter23117(t, "actual-origin23117.json.gz", out.Bytes())
	if code != exitOK {
		t.Fatal("actual saved record", code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	input, runtime := receipt.OriginInput, receipt.Result.Runtime
	if receipt.Decision != "OBSERVED" || input.ProducerExecutions != 0 || input.ProducerModelCalls != 0 || input.ReturnedRows != 9 || input.TotalRows != 27 || !input.Truncated || runtime.FiniteTotal != 0 || runtime.ModelCalls != 0 || len(runtime.Traces) != 9 {
		t.Fatal("original producer remains recorded;page9/27;input-only finite0", input)
	}
	counts := map[string]int{}
	for _, trace := range runtime.Traces {
		var actual struct {
			Status     string `json:"status"`
			FieldID    string `json:"field_id"`
			Expression string `json:"expression"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || !strings.HasPrefix(actual.FieldID, "urn:assembledlength22133:") {
			t.Fatal("original field identity", trace.Deliveries)
		}
		counts[actual.Status]++
	}
	if counts["DECLARED_ALTERNATIVE"] != 3 || counts["FIELD_TARGET"] != 3 || counts["UNRESOLVED"] != 3 {
		t.Fatal("source-owned starter interpretation", counts)
	}
	t.Log("practice4/4;relocated replay1 input without finite score;actual saved origins9/27:3 alternative,3 target,3 unresolved;original producer/model0")
}

func saveOriginsStarter23117(t *testing.T, name string, raw []byte) {
	t.Helper()
	root := os.Getenv("GOOO_FRESH_CLI_OUTPUT")
	if root == "" {
		return
	}
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	z := gzip.NewWriter(f)
	if _, err := z.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
