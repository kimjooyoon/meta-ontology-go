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

func TestGoooRecordChoiceOrigins22291(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/choice22133-runtime.json.gz"))
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
	var saved bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "record22291.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	rows, selected, err := bodyexecution.SelectRecordedNativeOrigins(saved, bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Limit: 9})
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "field-input-locations", "--origin-limit", "9", filepath.Join(root, "examples/record-choice-origins/gooo.workspace.json")}
	code := run(args, &out, &diagnostic)
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		file, err := os.Create(filepath.Join(output, "reader22291.json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		z := gzip.NewWriter(file)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
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
	r := receipt.Result.Runtime
	input := receipt.OriginInput
	if receipt.Decision != "OBSERVED" || r.FiniteTotal != 0 || r.ModelCalls != 0 || input.ProducerExecutions != 0 || input.ProducerModelCalls != 0 || input.FieldFileSourcesSHA256 != saved.FieldFileSources.SHA256 || len(rows) != 9 || len(r.Traces) != len(rows) || selected.TotalRows != 27 || !input.Truncated {
		t.Fatal("record-only page and unchanged producer scope", input)
	}
	choices, targets, unresolved := 0, 0, 0
	for i, trace := range r.Traces {
		var original bodyexecution.CompositionFieldFileFact
		if err := json.Unmarshal(rows[i], &original); err != nil {
			t.Fatal(err)
		}
		var actual struct {
			Status     string `json:"status"`
			Selected   bool   `json:"selected_alternative"`
			Mapped     bool   `json:"mapped"`
			Filename   string `json:"filename"`
			Start      int    `json:"source_start"`
			End        int    `json:"source_end"`
			Expression string `json:"expression"`
			RecordID   string `json:"record_id"`
			FieldID    string `json:"field_id"`
			Assignment int    `json:"assignment_index"`
			ActivityID string `json:"original_activity_id"`
			Reason     string `json:"reason"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("Gooo original choice output")
		}
		status := "UNRESOLVED"
		if original.Category == "record_choice" {
			status = "DECLARED_ALTERNATIVE"
			choices++
		} else if original.Category == "target" {
			status = "FIELD_TARGET"
			targets++
		} else {
			unresolved++
		}
		if actual.Status != status || actual.Selected != (original.Category == "record_choice") || actual.Mapped != (original.Mapping == "MAPPED") || actual.Filename != original.Filename || actual.Start != original.SourceStart || actual.End != original.SourceEnd || actual.Expression != original.Expression || actual.RecordID != original.RecordID || actual.FieldID != original.FieldID || actual.Assignment != original.AssignmentIndex || actual.ActivityID != original.OriginalActivityID || actual.Reason != original.Reason {
			t.Fatal("Gooo declared source policy", original, actual)
		}
	}
	if choices != 3 || targets != 3 || unresolved != 3 {
		t.Fatal("page populations", choices, targets, unresolved)
	}
	t.Log("real CLI/Gooo9/9;three declared alternatives,three field targets,three unresolved template values;page9/27;producer executions0;model0;input-only finite0")
}
