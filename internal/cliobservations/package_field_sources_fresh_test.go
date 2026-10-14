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

func TestGoooFieldSourceLocations16411(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/field-files16301-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	var r bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "saved16411.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	compared, mapped, unmapped := 0, 0, 0
	for page, offset := range []int{0, 12, 2 * len(r.Traces[0].FieldWrites)} {
		q := bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Offset: offset, Limit: 12}
		rows, _, err := bodyexecution.SelectRecordedNativeOrigins(r, q)
		if err != nil {
			t.Fatal(err)
		}
		var out, diagnostic bytes.Buffer
		args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-kind", "field-input-locations", "--origin-offset", strconv.Itoa(offset), "--origin-limit", "12", filepath.Join(root, "examples/field-source-map/gooo.workspace.json")}
		if code := run(args, &out, &diagnostic); code != exitOK {
			t.Fatal(code, diagnostic.String(), out.String())
		}
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
			t.Fatal(err, receipt.Error)
		}
		if receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 || receipt.OriginInput.FieldSourcesSHA256 != r.FieldSources.SHA256 || receipt.OriginInput.FieldFileSourcesSHA256 != r.FieldFileSources.SHA256 {
			t.Fatal("record-only source query scope", receipt.OriginInput)
		}
		if len(rows) != len(receipt.Result.Runtime.Traces) {
			t.Fatal("query/result denominator")
		}
		for i, trace := range receipt.Result.Runtime.Traces {
			var original bodyexecution.CompositionFieldFileFact
			if err := json.Unmarshal(rows[i], &original); err != nil {
				t.Fatal(err)
			}
			var actual struct {
				Status     string `json:"status"`
				Kind       string `json:"anchor_kind"`
				Mapped     bool   `json:"mapped"`
				Filename   string `json:"filename"`
				Start      int    `json:"source_start"`
				End        int    `json:"source_end"`
				Expression string `json:"expression"`
				Field      string `json:"field_id"`
				Assignment int    `json:"assignment_index"`
			}
			if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
				t.Fatal("Gooo location response")
			}
			status, kind := "UNRESOLVED", "FIELD_VALUE"
			if original.Mapping == "MAPPED" {
				status = "ORIGINAL_FILE"
				mapped++
			} else {
				unmapped++
			}
			if original.Category == "target" {
				kind = "FIELD_TARGET"
			}
			if actual.Status != status || actual.Kind != kind || actual.Mapped != (original.Mapping == "MAPPED") || actual.Filename != original.Filename || actual.Start != original.SourceStart || actual.End != original.SourceEnd || actual.Expression != original.Expression || actual.Field != original.FieldID || actual.Assignment != original.AssignmentIndex {
				t.Fatal("Gooo source policy", original, actual)
			}
			compared++
		}
		if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
			f, e := os.Create(filepath.Join(dir, "reader16411-page"+strconv.Itoa(page)+".json.gz"))
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
	}
	if compared != 36 || mapped == 0 || unmapped == 0 {
		t.Fatal("coverage", compared, mapped, unmapped)
	}
	t.Logf("Gooo original-file inspection%d/36;mapped%d unresolved%d;target/value scopes;three distinct pages;producer executions0;model0;input-only finite0", compared, mapped, unmapped)
}
