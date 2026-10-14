package main

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

func TestGoooFieldLocalBridge15317(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	example := filepath.Join(root, "examples/local-field-bridge")
	file, err := os.Open(filepath.Join(example, "recorded-runtime.json.gz"))
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
	var recorded bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "saved15317.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	offset, compared := 0, 0
	for c, trace := range recorded.Traces {
		for i, use := range trace.FieldUses {
			site := recorded.FieldUseSites[use.SiteIndex]
			binding := recorded.FieldBindings[site.Binding]
			if binding.Name != "forwarded" || binding.ActivityID != "urn:gooo:local-bridge:build" {
				continue
			}
			var out, diagnostic bytes.Buffer
			args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-shape", "field-local-columns", "--origin-offset", strconv.Itoa(offset + i), filepath.Join(example, "trace.workspace.json")}
			if code := run(args, &out, &diagnostic); code != exitOK {
				t.Fatal("Gooo mixed-journal traversal", c, i, code, diagnostic.String(), out.String())
			}
			var receipt packageExecutionReceipt
			if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
				t.Fatal(err, receipt.Error)
			}
			if receipt.Decision != "OBSERVED" || receipt.Result.Runtime.ModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 || receipt.OriginInput.InputShape != "gooo/field-local-lineage-columns/v1" || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 || receipt.OriginInput.BridgeObservation == nil {
				t.Fatal("record-only transport scope", receipt.OriginInput)
			}
			var got struct {
				Status      string `json:"status"`
				Roots       string `json:"roots"`
				RootCount   int    `json:"root_count"`
				Carrier     int    `json:"carrier_writes"`
				FieldReads  int    `json:"field_read_edges"`
				LocalReads  int    `json:"local_read_edges"`
				Unresolved  int    `json:"unresolved_links"`
				Shared      int    `json:"shared_rhs_writes"`
				RecordingOK bool   `json:"recording_complete"`
			}
			traces := receipt.Result.Runtime.Traces
			if len(traces) != 1 || len(traces[0].Deliveries) != 1 || json.Unmarshal(traces[0].Deliveries[0].Actual, &got) != nil {
				t.Fatal("mixed Gooo result shape")
			}
			input, minimum := "tag", 2
			if binding.Record.Fields[site.Field].ID == "urn:gooo:local-bridge:score" {
				input = "left"
				if c%2 == 1 {
					input, minimum = "right", 1
				}
			} else if binding.Record.Fields[site.Field].ID == "urn:gooo:local-bridge:live" {
				input = "enabled"
			}
			want := "urn:gooo:local-bridge:build|input|urn:gooo:local-bridge:input-" + input + ";"
			if got.Status != "SCOPED_FIELD_LOCAL_ORIGINS" || got.Roots != want || got.RootCount != 1 || got.Carrier < minimum || got.FieldReads < 1 || got.LocalReads < 1 || got.Unresolved != 0 || got.Shared != 0 || !got.RecordingOK {
				t.Fatal("Gooo-owned historical scalar carrier origins", c, i, got, want)
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
				if err := os.WriteFile(filepath.Join(dir, "consumer15317-row"+strconv.Itoa(compared)+".json.gz"), compressed.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
			}
			compared++
		}
		offset += len(trace.FieldUses)
	}
	if compared != 12 {
		t.Fatal("expected twelve distinct saved field queries", compared)
	}
	t.Log("12 saved mixed-journal queries->real32-column Gooo traversal12/12;integer/bool/string historical scalar carriers;equal-zero paths;original producer executions0;models0;input-only finite0")
}
