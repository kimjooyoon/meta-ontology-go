package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceCollectionFilterImportedPredicate(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"}, Packages: []packageruntime.PackageSpec{
		{Path: "example/core", Name: "core", Sources: []packageruntime.Source{{Filename: "keep.gooo", Content: "package core\nnamespace core\nentity Text id \"urn:gooo:type:string\"\nentity Boolean id \"urn:gooo:type:boolean\"\nactivity Keep(Text) -> Boolean id \"urn:filter-import:keep\" computes `return len(input)>0`\n"}}},
		{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{{Filename: "main.gooo", Content: "package app\nnamespace app\nimport rules \"example/core\"\nentity Input id \"urn:filter-import:input\" fields {field values id \"urn:filter-import:input-values\" type string required many}\nactivity Main(Input) -> Input id \"urn:filter-import:main\" computes `return Input{values:filter_values(input.values,rules.Keep)}`\n"}}},
	}}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{
		{Inputs: map[string]json.RawMessage{"example/app:Main": json.RawMessage(`{"values":["","ok","ok","한글",""]}`)}, Expected: map[string]json.RawMessage{"example/app:Main": json.RawMessage(`{"values":["ok","ok","한글"]}`)}},
	}}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 1 || !result.Runtime.RuntimeReplayed {
		t.Fatal(err, result.Runtime.Failure)
	}
	if len(result.Program.PureCalls.Sites) != 1 || len(result.Runtime.ConstructionSites) != 1 || result.Runtime.ConstructionSites[0].PredicateActivityID != "urn:filter-import:keep" {
		t.Fatal("imported predicate binding", result.Program.PureCalls, result.Runtime.ConstructionSites)
	}
	if len(result.Runtime.Traces) != 1 || len(result.Runtime.Traces[0].Constructions) != 1 || len(result.Runtime.Traces[0].Constructions[0].Elements) != 3 {
		t.Fatal("imported filter origins", result.Runtime.Traces)
	}
	if _, err := ReplayWorkspace(context.Background(), manifest, result, suite, ""); err != nil {
		t.Fatal("saved imported filter replay", err)
	}
	manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content, `len(input)>0`, `len(input)==0`, 1)
	if _, err := ReplayWorkspace(context.Background(), manifest, result, suite, ""); err == nil {
		t.Fatal("changed predicate reused receipt")
	}
}
