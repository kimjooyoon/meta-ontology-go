package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func recordedCriteriaRejection(t *testing.T) bodyConstructOutput {
	t.Helper()
	source := strings.Replace(cliCriteriaSource, `alternative "0-input"`, `alternative "input+0"`, 1)
	source = strings.Replace(source, `attempts "2"`, `observe { criteria_input "[11]" max_candidates "2" max_rounds "1" } attempts "2"`, 1)
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "report.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	c := bodyexecution.JointConstruction{Initial: bodyexecution.Composition{Steps: []bodyexecution.CompositionStep{{Generation: r}}}}
	return bodyConstructOutput{GeneratedNow: true, Construction: c}
}

func TestCriteriaRejectionSavedReportUsesRecordedReasonsOnly(t *testing.T) {
	saved := recordedCriteriaRejection(t)
	raw, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	for _, format := range []string{"text", "markdown", "json"} {
		var out, diagnostic bytes.Buffer
		if code := run([]string{"body-construct", "--report", file, "--format", format}, &out, &diagnostic); code != exitOK {
			t.Fatal(code, diagnostic.String())
		}
		if format == "json" {
			if !bytes.Equal(raw, out.Bytes()) {
				t.Fatal("derived explanation changed the saved JSON")
			}
			continue
		}
		for _, want := range []string{"Initial Gooo rejection explanation", "ALL_DECLARED_CANDIDATES_REJECTED", "2 rejected, 2 explained",
			"negative (Negative) returned false", "candidate 0 produced", "candidate 1 produced", "0 model calls and 0 program executions",
			"Recorded claims have not been reverified", "other implementations may satisfy"} {
			if !strings.Contains(html.UnescapeString(out.String()), want) {
				t.Fatalf("missing %q: %s", want, out.String())
			}
		}
	}
	unchanged, _ := os.ReadFile(file)
	if !bytes.Equal(raw, unchanged) {
		t.Fatal("saved observation was changed")
	}
}
