package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func exampleGraph(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../openfga-model-change/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := bodycodegen.InspectTypedValueOrigins("comparison.gooo", source, "Compare")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := project(r, "gooo://authorization-delta/comparison/model-match", true, true)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestExampleQueryUsesCompleteSourceGraph(t *testing.T) {
	var out bytes.Buffer
	err := run(context.Background(), []string{"--activity", "Compare", "--field", "gooo://authorization-delta/comparison/model-match", "--include-control", "--include-history", "../openfga-model-change/main.gooo"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	var result queryResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var summary struct {
		Status       string
		Contributors int
		Reachable    int
		Versions     int
		Total        int
		Paths        string
		SourceSHA    string `json:"source_sha256"`
		FlowSHA      string `json:"flow_sha256"`
		ReportSHA    string `json:"report_sha256"`
	}
	if err := json.Unmarshal(result.Result, &summary); err != nil {
		t.Fatal(err)
	}
	if len(result.ActivityProjections) != 3 || result.Steps != 44 || result.Budget != 44 || summary.Status != "RESOLVED" || summary.Contributors != 27 || summary.Total != 446 || summary.Reachable != 368 || summary.Versions != 74 || result.ReaderSourceSHA256 != sha(readerSource) || summary.SourceSHA == "" || summary.FlowSHA == "" || summary.ReportSHA == "" {
		t.Fatal("lost work, full input or original identities", out.String())
	}
	// This field affects source control. It is independently named here rather
	// than copied from the Gooo-produced summary or a normalized projection.
	if !strings.Contains(summary.Paths, "input:gooo://authorization-delta/pair/after > gooo://authorization-delta/observation/response;") {
		t.Fatal("control source path missing", summary.Paths)
	}
}

func TestExampleQueryRejectsInvalidGraphFactsWithoutPartialPaths(t *testing.T) {
	original := exampleGraph(t)
	for _, kind := range []string{"page", "value", "node", "bit"} {
		t.Run(kind, func(t *testing.T) {
			var g originGraph
			if err := json.Unmarshal(original, &g); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "page":
				g.Pages[0].First = 2
			case "value":
				g.ValuePages[0].Values[0].Roots[0] = 0
			case "node":
				g.Pages[0].Nodes[10].A = g.Pages[0].Nodes[10].NodeID
			case "bit":
				g.Pages[0].Nodes[10].Bit = 3
			}
			raw, _ := json.Marshal(g)
			result, err := query(context.Background(), readerSource, raw)
			if err != nil {
				t.Fatal(err)
			}
			var summary struct {
				Status       string
				Paths        string
				Contributors int
			}
			if err := json.Unmarshal(result.Result, &summary); err != nil {
				t.Fatal(err)
			}
			if summary.Status != "UNUSABLE_INPUT" || summary.Paths != "" || summary.Contributors != 0 {
				t.Fatal("invalid graph became an accepted partial answer", string(result.Result))
			}
		})
	}
}

func TestExampleQueryIdentityAndDeclaredTextBound(t *testing.T) {
	ctx := context.Background()
	original := exampleGraph(t)
	begin, err := bodycodegen.NewFixedProgram(ctx, "reader.gooo", readerSource, "Begin")
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := begin.Evaluate(ctx, tuple(original))
	if err != nil {
		t.Fatal(err)
	}
	step, err := bodycodegen.NewFixedProgram(ctx, "reader.gooo", readerSource, "Step")
	if err != nil {
		t.Fatal(err)
	}
	var g originGraph
	if err := json.Unmarshal(original, &g); err != nil {
		t.Fatal(err)
	}
	g.Control = false
	changed, _ := json.Marshal(g)
	next, err := step.Evaluate(ctx, tuple(changed, cursor.Value))
	if err != nil {
		t.Fatal(err)
	}
	state, err := cursorProgress(next.Value)
	if err != nil || state.Status != "QUERY_CHANGED" {
		t.Fatal("query reuse changed source edge meaning", state, err)
	}
	small := bytes.ReplaceAll(readerSource, []byte(`max_utf8_bytes "65536"`), []byte(`max_utf8_bytes "1024"`))
	if bytes.Equal(small, readerSource) {
		t.Fatal("source-owned Text declarations disappeared")
	}
	if _, err := query(ctx, small, original); err == nil || !strings.Contains(err.Error(), "1024") {
		t.Fatal("default-sized query concealed its actual Text refusal", err)
	}
}
