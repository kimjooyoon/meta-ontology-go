package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooRecordedFieldLineage14111(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := filepath.Join(sourceRoot, "examples/native-field-lineage")
	file, err := os.Open(filepath.Join(root, "recorded-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var runtime bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &runtime); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "recorded14111.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	offset := 0
	for c, trace := range runtime.Traces {
		selected := -1
		for i, use := range trace.FieldUses {
			site := runtime.FieldUseSites[use.SiteIndex]
			binding := runtime.FieldBindings[site.Binding]
			if binding.ActivityID == "urn:gooo:native-fields:read" && binding.Record.Fields[site.Field].ID == "urn:gooo:native-fields:right" {
				selected = i
			}
		}
		if selected < 0 {
			t.Fatal("recorded final field read absent", c)
		}
		var out, diagnostics bytes.Buffer
		args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "field-lineage-columns", "--origin-offset", strconv.Itoa(offset + selected), filepath.Join(root, "gooo.workspace.json")}
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal("Gooo field ancestry", c, code, diagnostics.String(), out.String())
		}
		if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
			var compressed bytes.Buffer
			writer := gzip.NewWriter(&compressed)
			if _, err := writer.Write(out.Bytes()); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "lineage14111-case"+strconv.Itoa(c)+".json.gz"), compressed.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Decision != "OBSERVED" || receipt.OriginInput == nil {
			t.Fatal(err)
		}
		if receipt.OriginInput.InputShape != "gooo/field-lineage-columns/v1" || receipt.OriginInput.ConsumerRows != 1 || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 {
			t.Fatal("recorded-consumer scope", receipt.OriginInput)
		}
		var result struct {
			Status            string `json:"status"`
			Roots             string `json:"roots"`
			RootCount         int    `json:"root_count"`
			Reachable         int    `json:"reachable_writes"`
			Parents           int    `json:"parent_links"`
			Reads             int    `json:"field_read_edges"`
			Unresolved        int    `json:"unresolved_links"`
			Shared            int    `json:"shared_rhs_writes"`
			RecordingComplete bool   `json:"recording_complete"`
		}
		traces := receipt.Result.Runtime.Traces
		if len(traces) != 1 || len(traces[0].Deliveries) != 1 || json.Unmarshal(traces[0].Deliveries[0].Actual, &result) != nil {
			t.Fatal("Gooo ancestry result shape")
		}
		field := "right"
		if c%2 == 1 {
			field = "left"
		}
		want := "urn:gooo:native-fields:build|input|urn:gooo:native-fields:input-" + field + ";"
		if result.Status != "SCOPED_FIELD_ORIGINS" || result.Roots != want || result.RootCount != 1 || result.Reachable < 5 || result.Parents < 3 || result.Reads < 1 || result.Unresolved != 0 || result.Shared != 0 || !result.RecordingComplete {
			t.Fatal("Gooo-owned transitive field origins", c, result, want)
		}
		offset += len(trace.FieldUses)
	}
	t.Log("4 distinct recorded reads->real Gooo ancestry traversal4/4;copy/helper/record-return/atomic-swap paths and equal-zero cases;original producer executions0;input-only finite0;models0")
}
