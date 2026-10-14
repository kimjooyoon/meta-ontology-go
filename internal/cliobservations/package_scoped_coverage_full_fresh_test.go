package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooScopedOriginCoverage19511(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodycodegen/testdata/counterfill18717.json.gz"))
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
	var saved struct {
		Runtime bodyexecution.CompositionRuntime `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	recorded, err := json.Marshal(saved.Runtime)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "counter-origins19511.json")
	if err := os.WriteFile(path, recorded, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "local-fill-fragments", "--origin-shape", "local-fill-columns", "--origin-offset", "0", "--origin-limit", "128", filepath.Join(root, "examples/scoped-origin-coverage/gooo.workspace.json")}
	if code := run(args, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	rows, _, err := bodyexecution.SelectRecordedNativeOriginPage(saved.Runtime, bodyexecution.NativeOriginQuery{Kind: "local-fill-fragments", Offset: 0, Limit: 128})
	if err != nil {
		t.Fatal(err)
	}
	candidates, templates, grammars, mapped := 0, 0, 0, 0
	for _, row := range rows {
		var fact bodyexecution.LocalFillFragmentFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.Category == "declared_candidate" {
			candidates++
		}
		if fact.Category == "retained_template" {
			templates++
		}
		if fact.Category == "derived_grammar" {
			grammars++
		}
		if fact.SourceSHA != "" && fact.SourceStart >= 0 && fact.SourceEnd > fact.SourceStart {
			mapped++
		}
	}
	var actual struct {
		Status     string `json:"status"`
		Rows       int    `json:"fragment_rows"`
		Mapped     int    `json:"mapped_rows"`
		Missing    int    `json:"missing_rows"`
		Templates  int    `json:"template_rows"`
		Candidates int    `json:"candidate_rows"`
		Grammars   int    `json:"grammar_rows"`
		Ratio      int    `json:"row_mapping_basis_points"`
		Whole      bool   `json:"whole_recording"`
	}
	traces := receipt.Result.Runtime.Traces
	if len(traces) != 1 || len(traces[0].Deliveries) != 1 || json.Unmarshal(traces[0].Deliveries[0].Actual, &actual) != nil {
		t.Fatal("Gooo coverage response")
	}
	if len(rows) == 0 || actual.Status != "RECORDED_FRAGMENTS" || actual.Rows != len(rows) || actual.Mapped != mapped || actual.Missing != len(rows)-mapped || actual.Templates != templates || actual.Candidates != candidates || actual.Grammars != grammars || actual.Ratio != mapped*10000/len(rows) || !actual.Whole {
		t.Fatal("Gooo row coverage", actual, len(rows), mapped, candidates, templates, grammars)
	}
	if receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 {
		t.Fatal("saved input scope")
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		f, e := os.Create(filepath.Join(root, "coverage19511.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if _, e = z.Write(out.Bytes()); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("new Gooo for-let metaprogram:actual saved fragment rows%d,mapped%d,candidates%d,templates%d,grammars%d;row-only ratio and whole-recording scope;original producer executions0;model0;input-only finite0", actual.Rows, actual.Mapped, actual.Candidates, actual.Templates, actual.Grammars)
}
