package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestPackageDerivedOrigin11903(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	dir := filepath.Join(root, "examples/derived-origin")
	checkGrammarRoundtrip11903(t, filepath.Join(dir, "main.gooo"))
	var producer, consumer, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(dir, "cases.json"), filepath.Join(dir, "gooo.workspace.json")}, &producer, &diagnostics); code != exitOK {
		t.Fatal("derived producer", code, diagnostics.String(), producer.String())
	}
	saveGrammarOutput11903(t, "producer11903.json", producer.Bytes())
	receipt := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(receipt, producer.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	reader := filepath.Join(root, "examples/native-origin-reader/local-fill/reader.workspace.json")
	if code := run([]string{"package", "execute", "--json", "--origin-receipt", receipt, "--origin-kind", "local-fill-fragments", "--origin-limit", "6", reader}, &consumer, &diagnostics); code != exitOK {
		t.Fatal("actual grammar reader", code, diagnostics.String(), consumer.String())
	}
	saveGrammarOutput11903(t, "consumer11903.json", consumer.Bytes())
	checkGrammarReceipt11903(t, producer.Bytes(), consumer.Bytes())
	t.Log("source grammar and assembly roundtrip retained; fresh CLI producer3/3; actual derived/template rows6/6 classifications matched; grammar positions exact; no fabricated expression position; read producer0/models0")
}

func checkGrammarRoundtrip11903(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	formatted, err := syntax.Format(file)
	if err != nil {
		t.Fatal(err)
	}
	before, err := bodycodegen.SourceAssembly(context.Background(), path, raw, "Prepare")
	if err != nil {
		t.Fatal(err)
	}
	after, err := bodycodegen.SourceAssembly(context.Background(), path, []byte(formatted), "Prepare")
	if err != nil {
		t.Fatal(err)
	}
	a, err := before.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	b, err := after.Canonical()
	if err != nil || a != b {
		t.Fatal("formatting changed grammar or assembly intent", err)
	}
}

func checkGrammarReceipt11903(t *testing.T, producer, consumer []byte) {
	t.Helper()
	var p, c packageExecutionReceipt
	if err := json.Unmarshal(producer, &p); err != nil || p.Result == nil || p.Result.Runtime.FinitePassed != 3 || p.Result.Runtime.ModelCalls != 0 {
		t.Fatal("grammar producer population", err)
	}
	if err := json.Unmarshal(consumer, &c); err != nil || c.Result == nil || c.OriginInput == nil || c.OriginInput.ReturnedRows != 6 || c.OriginInput.ProducerExecutions != 0 || c.Result.Runtime.ModelCalls != 0 {
		t.Fatal("grammar consumer population", err)
	}
	rows, _, err := bodyexecution.SelectRecordedNativeOrigins(p.Result.Runtime, bodyexecution.NativeOriginQuery{Kind: "local-fill-fragments", Limit: 6})
	if err != nil || len(rows) != 6 {
		t.Fatal(err, len(rows))
	}
	for index, row := range rows {
		var fact bodyexecution.LocalFillFragmentFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		want := "template"
		if fact.AssignmentIndex == 0 {
			want = "grammar"
			if fact.Category != "derived_grammar" || string([]byte(p.Result.Program.Source)[fact.SourceStart:fact.SourceEnd]) != "integer-offset-constant/v1" {
				t.Fatal("derived origin did not locate its declared grammar", fact)
			}
		}
		trace := c.Result.Runtime.Traces[index]
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want {
			t.Fatal("Gooo grammar classification", index, trace)
		}
	}
	for _, site := range p.Result.Runtime.LocalSites {
		origin := site.Source.Construction
		if origin == nil || origin.CandidateSource != nil {
			t.Fatal("derived origin claimed an authored candidate", site)
		}
		for _, fragment := range origin.Fragments {
			if fragment.Kind == "derived_grammar" && (fragment.GrammarSource == nil || fragment.ExpressionSource != nil) {
				t.Fatal("derived expression acquired an invented literal location", fragment)
			}
		}
	}
}

func saveGrammarOutput11903(t *testing.T, name string, raw []byte) {
	t.Helper()
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
