package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceCollectionMapImportedMapper(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"}, Packages: []packageruntime.PackageSpec{
		{Path: "example/core", Name: "core", Sources: []packageruntime.Source{{Filename: "validate.gooo", Content: "package core\nnamespace core\nentity Text id \"urn:gooo:type:string\"\nentity Boolean id \"urn:gooo:type:boolean\"\nactivity Valid(Text) -> Boolean id \"urn:map-import:valid\" computes `return input==\"ok\"`\n"}}},
		{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{{Filename: "main.gooo", Content: "package app\nnamespace app\nimport rules \"example/core\"\nentity Input id \"urn:map-import:input\" fields {field values id \"urn:map-import:input-values\" type string required many}\nentity Output id \"urn:map-import:output\" fields {field values id \"urn:map-import:output-values\" type boolean required many}\nactivity Main(Input) -> Output id \"urn:map-import:main\" computes `return Output{values:map_values(input.values,rules.Valid)}`\n"}}},
	}}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{
		{Inputs: map[string]json.RawMessage{"example/app:Main": json.RawMessage(`{"values":["ok","","ok","한글"]}`)}, Expected: map[string]json.RawMessage{"example/app:Main": json.RawMessage(`{"values":[true,false,true,false]}`)}},
	}}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 1 || !result.Runtime.RuntimeReplayed {
		t.Fatal(err, result.Runtime.Failure)
	}
	if len(result.Program.PureCalls.Sites) != 1 || len(result.Runtime.ConstructionSites) != 1 || result.Runtime.ConstructionSites[0].MapperActivityID != "urn:map-import:valid" {
		t.Fatal("import mapping", result.Program.PureCalls)
	}
	if _, err := ReplayWorkspace(context.Background(), manifest, result, suite, ""); err != nil {
		t.Fatal("saved imported map replay", err)
	}
	manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content, `input=="ok"`, `input!="ok"`, 1)
	if _, err := ReplayWorkspace(context.Background(), manifest, result, suite, ""); err == nil {
		t.Fatal("changed mapper reused receipt")
	}
}
