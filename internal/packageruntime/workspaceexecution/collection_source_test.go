package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceCollectionCallsAndSource7379(t *testing.T) {
	body := `let 합계 = 0; for 값 in input.amounts { let 조정 = 계산.Adjust(값); if 조정 > 0 { 합계 = 합계 + 조정 } }; return Summary{total: 합계, amounts: input.amounts}`
	app := `package app
namespace workspace7379_app
import 계산 "collection7379/math"
entity Basket id "urn:workspace7379:basket" fields {
 field amounts id "urn:workspace7379:amounts" type integer required many
 field flags id "urn:workspace7379:flags" type boolean optional many
}
entity Summary id "urn:workspace7379:summary" fields {
 field total id "urn:workspace7379:total" type integer required one
 field amounts id "urn:workspace7379:originals" type integer required many
}
activity Compute(Basket) -> Summary id "urn:workspace7379:compute" computes ` + strconv.Quote(body) + "\n"
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "collection7379/app", Activity: "Compute"},
		Packages: []packageruntime.PackageSpec{
			{Path: "collection7379/math", Name: "math", Sources: []packageruntime.Source{{Filename: "보정.gooo", Content: `package math
namespace workspace7379_math
entity Integer id "urn:gooo:type:integer"
activity Adjust(Integer) -> Integer id "urn:workspace7379:adjust" computes "return input + 2"
`}}},
			{Path: "collection7379/app", Name: "app", Imports: []string{"collection7379/math"},
				Sources: []packageruntime.Source{{Filename: "계산.gooo", Content: app}}},
		}}
	before, _ := json.Marshal(manifest)
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"collection7379/app:Compute":{"amounts":[3,-8,20]}},"expected":{"collection7379/app:Compute":{"total":27,"amounts":[3,-8,20]}}},
{"inputs":{"collection7379/app:Compute":{"amounts":[],"flags":[]}},"expected":{"collection7379/app:Compute":{"total":0,"amounts":[]}}},
{"inputs":{"collection7379/app:Compute":{"amounts":[9007199254741001],"flags":[true,false]}},"expected":{"collection7379/app:Compute":{"total":9007199254741003,"amounts":[9007199254741001]}}}
]}`
	suite, err := bodyexecution.DecodeCompositionCases([]byte(cases))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal("collection workspace execution", err)
	}
	if result.Runtime.FinitePassed != 3 || result.Runtime.FiniteTotal != 3 || result.Runtime.ModelCalls != 0 || len(result.Runtime.ConditionSites) != 1 {
		t.Fatal("native collection workspace results", result.Runtime)
	}
	anchor := result.Runtime.ConditionSites[0].Source
	if anchor == nil || anchor.Status != "MAPPED" || anchor.Input == nil || anchor.Input.Status != "MAPPED" ||
		anchor.Input.Filename != "계산.gooo" || app[anchor.Input.Start:anchor.Input.End] != "조정 > 0" {
		t.Fatal("condition after imported helper call must cite original Korean file", anchor)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("workspace source changed during execution")
	}
	if directory := os.Getenv("GOOO_COLLECTION_OBSERVATION"); directory != "" {
		raw, err := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": json.RawMessage(cases), "result": result,
			"historical_source_executions": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "workspace7379.json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("7379: imported activity per element, record constructor after for-in, optional Boolean arrays, native3/3 and original file condition mapping; model0")
}
