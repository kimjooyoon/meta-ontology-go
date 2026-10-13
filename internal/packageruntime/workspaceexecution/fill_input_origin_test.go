package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestImportedFillInputOrigins4051(t *testing.T) {
	const dependency = `package rates
namespace rates4051
entity Integer id "urn:gooo:type:integer"
activity Check(Integer) -> Integer id "urn:rates4051:check" computes "return input + 17"
`
	const source = `package app
namespace app4051
import 資料 "origin4051/rates"
activity Check(資料.Integer) -> 資料.Integer id "urn:app4051:check" computes "if 資料.Check(input) > 71 && __GOOO_BODY_HOLE_gate__ { return input + 5 }; return input - 7" assembling {
 source_fill intent "元の条件と候補の出所を保つ" {
  hole "gate"
  candidate "qualified4051" { fill "gate" "資料.Check(input) > 93" }
  candidate "closed4051" { fill "gate" "資料.Check(input) > 193" }
 }
 case "79" -> "84"
 case "23" -> "16"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "origin4051/app", Activity: "Check"},
		Packages: []packageruntime.PackageSpec{
			{Path: "origin4051/rates", Name: "rates", Sources: []packageruntime.Source{{Filename: "기준.gooo", Content: dependency}}},
			{Path: "origin4051/app", Name: "app", Imports: []string{"origin4051/rates"}, Sources: []packageruntime.Source{{Filename: "판단.gooo", Content: source}}},
		}}
	before, _ := json.Marshal(manifest)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin4051/app:Check":81},"expected":{"origin4051/app:Check":86}},{"inputs":{"origin4051/app:Check":45},"expected":{"origin4051/app:Check":38}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("GOOO_FILL_INPUT_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "result": result}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "import4051.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if result.Runtime.FinitePassed != 2 || len(result.Runtime.ConditionSites) != 1 {
		t.Fatal("native origin result", result.Runtime)
	}
	link := result.Runtime.ConditionSites[0].Source.Construction
	if link == nil || link.Input == nil {
		t.Fatal("missing original package source")
	}
	origin := link.Input
	if origin.Status != "MAPPED" || origin.Filename != "판단.gooo" || origin.ActivityID != "urn:app4051:check" ||
		origin.SourceSHA256 != sourceSHA256([]byte(source)) || origin.CandidateID != "qualified4051" || len(origin.Fragments) != 2 {
		t.Fatal("original declaration identity differs", origin)
	}
	for _, fragment := range origin.Fragments {
		if fragment.Status != "MAPPED" || fragment.TemplateSource == nil {
			t.Fatal("original fragment unavailable", fragment)
		}
		if fragment.Kind == "declared_candidate" {
			if fragment.Expression != "資料.Check(input) > 93" || fragment.ExpressionSource == nil ||
				source[fragment.ExpressionSource.Start:fragment.ExpressionSource.End] != fragment.Expression {
				t.Fatal("authored candidate lost", fragment)
			}
		} else if !strings.Contains(source[fragment.TemplateSource.Start:fragment.TemplateSource.End], "資料.Check(input)") {
			t.Fatal("retained template lost its import spelling", fragment)
		}
	}
	raw, _ := json.Marshal(result)
	var saved Result
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		t.Fatal("saved origin cannot be decoded", err)
	}
	if _, _, err := reconstructWorkspace(context.Background(), manifest, saved); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) || result.Runtime.ModelCalls != 0 {
		t.Fatal("source or model boundary changed")
	}
	if _, ok := undoFillSourceEdits(bodycodegen.FillOriginRange{Start: 3, End: 5}, []BodySourceEdit{{Start: 2, End: 4, Replacement: "longer"}}); ok {
		t.Fatal("boundary inside rewritten name fabricated an original location")
	}
	t.Log("import4051: native2/2; imported candidate and retained template point to original Unicode file; strict saved decoding and construction read; same-name activities stay separate; model0")
}
