package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestLengthAssemblyOrigins22133(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../.."
	}
	dir := filepath.Join(root, "examples", "assembled-length-summary")
	source, err := os.ReadFile(filepath.Join(dir, "summary.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "origin/assembledlength22133", Activity: "Summarize"},
		Packages: []packageruntime.PackageSpec{{Path: "origin/assembledlength22133", Name: "assembledlength", Sources: []packageruntime.Source{{Filename: "조립집계22133.gooo", Content: string(source)}}}}}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		file, saveErr := os.Create(filepath.Join(output, "length-assembly22133.json.gz"))
		if saveErr != nil {
			t.Fatal(saveErr)
		}
		zip := gzip.NewWriter(file)
		if saveErr := json.NewEncoder(zip).Encode(result); saveErr != nil {
			t.Fatal(saveErr)
		}
		if saveErr := zip.Close(); saveErr != nil {
			t.Fatal(saveErr)
		}
		if saveErr := file.Close(); saveErr != nil {
			t.Fatal(saveErr)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 3 || result.Runtime.FiniteTotal != 3 || result.Runtime.ModelCalls != 0 || !result.Runtime.RuntimeReplayed {
		t.Fatal("independent native cases", result.Runtime.FinitePassed, result.Runtime.FiniteTotal)
	}
	selections := 0
	for _, step := range result.Composition.Steps {
		receipt := step.Generation.Report.RecordAssembly
		if receipt == nil {
			continue
		}
		if receipt.SelectedMask != 7 || receipt.Passed != 1 || receipt.Total != 1 || receipt.Holdout == nil || receipt.Holdout.Passed != 1 || receipt.Holdout.Total != 1 {
			t.Fatal("source field choices, training and held-out scopes", receipt)
		}
		selections++
	}
	if selections != 1 {
		t.Fatal("source-owned record selection", selections)
	}
	if result.Runtime.FieldFileSources == nil {
		t.Fatal("original choice file sources absent")
	}
	choices := map[string]bool{}
	for _, binding := range result.Runtime.FieldFileSources.Bindings {
		if choice := binding.Choice; choice != nil {
			if choice.OptionID != "value_second" || choice.Value.Status != "MAPPED" || choice.Declaration.Status != "MAPPED" {
				t.Fatal("original choice identity", choice)
			}
			choices[choice.ChoiceID] = true
		}
	}
	if len(choices) != 3 {
		t.Fatal("three source-owned choices", choices)
	}
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(result.Runtime, bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Limit: 128})
	if err != nil || report.ProducerExecutions != 0 {
		t.Fatal("saved field source query", err, report)
	}
	alternatives := 0
	for _, row := range rows {
		var fact bodyexecution.CompositionFieldFileFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.Category == "record_choice" {
			if fact.Mapping != "MAPPED" || fact.Filename != "조립집계22133.gooo" || fact.CandidateID != "" || fact.HoleID != "" || string(source[fact.SourceStart:fact.SourceEnd]) != fact.Expression {
				t.Fatal("original record choice expression", fact)
			}
			alternatives++
		}
	}
	if alternatives != 9 {
		t.Fatal("selected declared alternatives have no original input locations", alternatives)
	}
	t.Log("three source-owned record field choices;selection1/1,held-out1/1,separate native3/3;original choice declarations retained;two-run native;model0")
}
