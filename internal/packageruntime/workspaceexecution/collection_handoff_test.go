package workspaceexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestCollectionHandoff8411(t *testing.T) {
	root := os.Getenv("GOOO_HANDOFF_SOURCE")
	if root == "" {
		root = filepath.Join("..", "..", "..", "examples", "batch-handoff")
	}
	source, err := os.ReadFile(filepath.Join(root, "main.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	qualified := strings.NewReplacer(`"Select"`, `"handoff8411:Select"`, `"Summarize"`, `"handoff8411:Summarize"`).Replace(string(raw))
	suite, err := bodyexecution.DecodeCompositionCases([]byte(qualified))
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "handoff8411", Activity: "Summarize"},
		Packages: []packageruntime.PackageSpec{{Path: "handoff8411", Name: "batchhandoff8411",
			Sources: []packageruntime.Source{{Filename: "작업전달.gooo", Content: string(source)}}}},
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal("construct and execute collection handoff", err)
	}
	if result.Runtime.FinitePassed != 14 || result.Runtime.FiniteTotal != 14 || result.Runtime.ModelCalls != 0 {
		t.Fatal("seven input rows, two observed activity outputs each", result.Runtime.FinitePassed, result.Runtime.FiniteTotal)
	}
	if len(result.Composition.Plan.Edges) != 1 || len(result.Composition.Steps) != 2 {
		t.Fatal("source bind did not become a two-activity plan", result.Composition.Plan)
	}
	if len(result.BodyFills) != 1 {
		t.Fatal("one retained workspace body fill required", len(result.BodyFills))
	}
	fill := result.BodyFills[0].Generation.Report.BodyFill
	if fill == nil || fill.SelectedCandidateID != "inclusive" || fill.TestCasesPassed != 3 || fill.TestCasesTotal != 3 || fill.HoldoutCasesPassed != 1 || fill.HoldoutCasesTotal != 1 {
		t.Fatal("construction cases and holdout remain separate", fill)
	}
	before, _ := json.Marshal(result)
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{Limit: 128})
	if err != nil || flow.Truncated || flow.Analysis != "PARTIAL" {
		t.Fatal("bounded collection origin view", err, flow)
	}
	boundCandidate := false
	for _, row := range flow.Rows {
		if row.Kind != "candidate" || row.CandidateID != "inclusive" {
			continue
		}
		for _, step := range row.Path {
			boundCandidate = boundCandidate || step.Via == "bind"
		}
		original := false
		for _, location := range row.Locations {
			if location.View == "workspace_input_file" {
				original = location.Filename == "작업전달.gooo" && location.SourceSHA256 == sourceSHA256(source) &&
					strings.Contains(string(source[location.Start:location.End]), "duration <= input.budget - spent")
			}
		}
		if !original {
			t.Fatal("selected condition lost its original source", row)
		}
	}
	if !boundCandidate {
		t.Fatal("downstream summary lost the selected prefix's candidate origin", flow)
	}
	after, _ := json.Marshal(result)
	if !bytes.Equal(before, after) {
		t.Fatal("origin inspection changed the execution receipt")
	}
	if output := os.Getenv("GOOO_HANDOFF_OBSERVATION"); output != "" {
		data, err := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": json.RawMessage(raw),
			"result": result, "origins": flow, "native_input_rows": 7, "expected_activity_outputs": 14,
			"historical_source_executions": 0, "model_calls": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "handoff8411.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8411: computed prefix slice -> bound activity -> count/sum; native14/14 across7 inputs; selection3/3 holdout1/1; original candidate source across bind; static collection origins PARTIAL; model0")
}
