package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestIndexedBridgeOriginCLI(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	source, err := os.ReadFile("../../examples/record-collections/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := os.ReadFile("../../examples/record-collections/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := bodyexecution.GenerateCompositionWithOptions(context.Background(), "items.gooo", source, suite, bodyexecution.CompositionOptions{EntryActivity: "Total"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := bodyexecution.ExecuteComposition(context.Background(), "items.gooo", source, plan, suite, "")
	if err != nil || r.FinitePassed != 4 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("fresh source producer", err, r.FinitePassed)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "recorded-items.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	checkIndexedBridgeCLIPage(t, saved, 4, []int{2, 2}, []int{4, 8}, []int{1, 1})
	checkIndexedBridgeCLIPage(t, saved, 5, []int{2, 3}, []int{8, 1}, []int{1, 0})
	t.Log("new CLI shape runs Gooo on4 repeated location queries, including cross-case page;producer stays4 original finite cases;input-only consumer finite0;producer query executions0/models0")
}

func checkIndexedBridgeCLIPage(t *testing.T, saved string, offset int, cases, masks, positions []int) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", saved,
		"--origin-shape", "field-local-columns-v2", "--origin-offset", strconv.Itoa(offset),
		"--origin-limit", "2", "../../examples/local-field-bridge/indexed.workspace.json"}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "OBSERVED" || runtime.FiniteTotal != 0 || runtime.ModelCalls != 0 ||
		receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.InputShape != "gooo/field-local-lineage-columns/v2" ||
		receipt.OriginInput.ConsumerRows != 2 || len(runtime.Traces) != 2 {
		t.Fatal("new shape inflated finite coverage or executed producer", receipt.OriginInput)
	}
	for i, trace := range runtime.Traces {
		var actual struct {
			Summary struct {
				Case   int    `json:"case_index"`
				Status string `json:"status"`
			}
			Low   int `json:"indexed_low"`
			High  int `json:"indexed_high"`
			Roots int `json:"indexed_roots"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil ||
			actual.Summary.Case != cases[i] || actual.Summary.Status != "SCOPED_INDEXED_FIELD_LOCAL_ORIGINS" ||
			actual.Low != masks[i] || actual.High != 0 || actual.Roots != 1 {
			t.Fatal("source-owned selection or original case changed", i, actual)
		}
		var input bodyexecution.IndexedBridgeLineageColumns
		if err := json.Unmarshal(trace.Deliveries[0].Input, &input); err != nil {
			t.Fatal(err)
		}
		ref := input.Elements.Refs[input.History.RootAssignment]
		found := false
		for j, read := range input.Elements.ReadIDs {
			if read == ref {
				found = input.Elements.Indices[j] == positions[i] && input.Elements.FieldIDs[j] == "urn:record-items:items"
			}
		}
		if !found {
			t.Fatal("original element position changed", i, input.Elements)
		}
	}
}
