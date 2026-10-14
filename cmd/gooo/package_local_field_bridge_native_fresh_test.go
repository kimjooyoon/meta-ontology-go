package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestNativeLocalFieldBridge15217(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	example := filepath.Join(root, "examples/local-field-bridge")
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(example, "cases.json"), filepath.Join(example, "gooo.workspace.json")}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, receipt.Error)
	}
	r := receipt.Result.Runtime
	if receipt.Decision != "PASS" || r.FinitePassed != 4 || !r.RuntimeReplayed || r.ModelCalls != 0 || r.BridgeObservation == nil || r.BridgeObservation.Status != "COMPLETE" || r.BridgeObservation.CrossReads < 12 || r.BridgeObservation.Interrupted != 0 {
		t.Fatal("native mixed journal outcomes", r.FinitePassed, r.BridgeObservation)
	}
	for c, trace := range r.Traces {
		matched := 0
		for _, use := range trace.FieldUses {
			site := r.FieldUseSites[use.SiteIndex]
			binding := r.FieldBindings[site.Binding]
			if binding.Name != "forwarded" || binding.ActivityID != "urn:gooo:local-bridge:build" {
				continue
			}
			field := binding.Record.Fields[site.Field].ID
			input := "tag"
			minimumLocals := 2
			if field == "urn:gooo:local-bridge:score" {
				input = "left"
				if c%2 == 1 {
					input = "right"
					minimumLocals = 1
				}
			} else if field == "urn:gooo:local-bridge:live" {
				input = "enabled"
			}
			got, localWrites := nativeBridgeRoots15217(r, c, use.Assignment)
			if !slices.Equal(got, []string{"urn:gooo:local-bridge:input-" + input}) || localWrites < minimumLocals {
				t.Fatal("field/local/field origin", c, field, got, localWrites)
			}
			matched++
		}
		if matched != 3 {
			t.Fatal("final three field reads absent", c, matched)
		}
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
		if err := os.WriteFile(filepath.Join(dir, "producer15217.json.gz"), compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("new Gooo source4/4 actual mixed tuples and exact2-run bridge replay;12 final field roots through scalar local writes;overwritten fields retain earlier saved origins;equal-zero branch paths differ;models0;external oracle only, Gooo mixed-journal traversal still pending")
}

func nativeBridgeRoots15217(r bodyexecution.CompositionRuntime, c, assignment int) ([]string, int) {
	t := r.Traces[c]
	seen := map[[2]int]bool{}
	roots := map[string]bool{}
	locals := 0
	var visit func(int, int)
	visit = func(kind, index int) {
		if index < 0 || seen[[2]int{kind, index}] {
			return
		}
		seen[[2]int{kind, index}] = true
		if kind == 0 {
			w := t.FieldWrites[index]
			site := r.FieldSites[w.SiteIndex]
			binding := r.FieldBindings[site.Binding]
			if w.ParentAssignment >= 0 {
				visit(0, w.ParentAssignment)
				return
			}
			if site.Kind == "parameter" && binding.InputKind == "root_input" {
				roots[binding.Record.Fields[site.Field].ID] = true
				return
			}
			for _, u := range t.FieldUses[w.UseStart:w.UseEnd] {
				visit(0, u.Assignment)
			}
			b := t.Bridge.Fields[index]
			for _, u := range t.LocalUses[b.OtherStart:b.OtherEnd] {
				visit(1, u.Assignment)
			}
		} else {
			locals++
			w, b := t.LocalWrites[index], t.Bridge.Locals[index]
			for _, u := range t.LocalUses[w.UseStart:w.UseEnd] {
				visit(1, u.Assignment)
			}
			for _, u := range t.FieldUses[b.OtherStart:b.OtherEnd] {
				visit(0, u.Assignment)
			}
		}
	}
	visit(0, assignment)
	result := make([]string, 0, len(roots))
	for root := range roots {
		result = append(result, root)
	}
	slices.Sort(result)
	return result, locals
}
