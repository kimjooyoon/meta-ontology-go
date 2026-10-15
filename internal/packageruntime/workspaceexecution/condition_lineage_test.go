package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceConditionLineageOriginalFiles1879(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "origin1879/app", Activity: "Price"},
		Packages: []packageruntime.PackageSpec{
			{Path: "origin1879/rules", Name: "rules", Sources: []packageruntime.Source{{Filename: "규칙1879.gooo", Content: `package rules
namespace rules1879
entity Ticket id "origin1879://ticket" fields {
 field units id "origin1879://units" type "urn:gooo:type:integer" required one
}
entity Boolean id "origin1879://boolean"
activity Eligible(Ticket) -> Boolean id "origin1879://eligible" computes "if input.units <= 31 { return true }; return false"
`}}},
			{Path: "origin1879/app", Name: "app", Imports: []string{"origin1879/rules"}, Sources: []packageruntime.Source{{Filename: "가격1879.gooo", Content: `package app
namespace app1879
import 규칙 "origin1879/rules"
entity Integer id "origin1879://integer"
activity Price(규칙.Ticket) -> Integer id "origin1879://price" computes "let adjusted = 규칙.Ticket{units: input.units + 2}; if 규칙.Eligible(adjusted) { return adjusted.units * 5 }; return input.units - 1"
`}}},
		}}
	before, _ := json.Marshal(manifest)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin1879/app:Price":{"units":12}},"expected":{"origin1879/app:Price":70}},{"inputs":{"origin1879/app:Price":{"units":41}},"expected":{"origin1879/app:Price":40}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("GOOO_CONDITION_OBSERVATION_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(result, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "workspace1879-result.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "workspace1879-manifest.json"), append(before, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if result.Runtime.FinitePassed != 2 || len(result.Runtime.ConditionSites) != 2 {
		t.Fatal("fresh workspace behavior differs", result.Runtime.FinitePassed)
	}
	contents := map[string]string{}
	for _, pkg := range manifest.Packages {
		for _, source := range pkg.Sources {
			contents[pkg.Path+"\x00"+source.Filename] = source.Content
		}
	}
	fields := 0
	for _, site := range result.Runtime.ConditionSites {
		selected := site.Source
		if selected == nil || selected.Status != "MAPPED" || selected.Input == nil || selected.Input.Status != "MAPPED" {
			t.Fatal("selected or original origin lost", site)
		}
		input := selected.Input
		original := contents[input.PackagePath+"\x00"+input.Filename]
		quoted := strconv.Quote(input.Expression)
		if input.SourceSHA256 != sourceSHA256([]byte(original)) || original[input.Start:input.End] != quoted[1:len(quoted)-1] {
			t.Fatal("original file byte range differs", input)
		}
		if input.Declaration == nil || input.Declaration.Status != "MAPPED" || input.Declaration.StableID != site.ActivityID || !strings.Contains(original[input.Declaration.Start:input.Declaration.End], "activity ") {
			t.Fatal("original semantic declaration missing", input)
		}
		for _, field := range selected.Fields {
			fields++
			if field.FieldID != "origin1879://units" || field.Declaration == nil || field.Declaration.Status != "MAPPED" || !strings.Contains(result.Composition.GoooSource[field.Declaration.Start:field.Declaration.End], "field units") {
				t.Fatal("field declaration origin missing", field)
			}
		}
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) || fields != 1 || result.Runtime.ModelCalls != 0 {
		t.Fatal("original source or typed reference changed", fields)
	}
	t.Log("workspace1879: original-file condition2/2 and activity declarations2/2; typed field declaration1/1; qualified call/constructor length changes retained; native2/2; model0")
}
