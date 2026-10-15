package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const namedFillSource = `package namedfill
namespace namedfill
entity Integer id "urn:gooo:type:integer"
activity Adjust(amount Integer) -> Integer id "urn:namedfill:adjust" computes "return amount + __GOOO_BODY_HOLE_delta__" assembling {
 source_fill intent "입력 금액에 13을 더한다" {
  hole "delta"
  candidate "offset" { fill "delta" "13" }
  candidate "unchanged" { fill "delta" "0" }
 }
 case "3" -> "16"
}
activity Main(원금 Integer, 공제 Integer) -> Integer id "urn:namedfill:main" computes "return Adjust(원금) - 공제"
`

func TestAuthoredInputFillAndSourcePathsFirst(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "named/fill", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{{Path: "named/fill", Name: "namedfill",
			Sources: []packageruntime.Source{{Filename: "fill.gooo", Content: namedFillSource}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"named/fill:Main.input0":67,"named/fill:Main.input1":19},"expected":{"named/fill:Main":61}},
 {"inputs":{"named/fill:Main.input0":-29,"named/fill:Main.input1":5},"expected":{"named/fill:Main":-21}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FiniteTotal != 2 || result.Runtime.FinitePassed != 2 || result.Runtime.ModelCalls != 0 {
		t.Fatal(result.Runtime)
	}
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{CandidateID: "offset", Limit: 32})
	if err != nil || flow.Analysis != "MAPPED" || len(flow.Rows) != 1 || flow.Rows[0].Kind != "candidate" ||
		flow.Rows[0].ActivityID != "urn:namedfill:adjust" || flow.Rows[0].HoleID != "delta" {
		t.Fatal("named source selection origin", flow, err)
	}
	if output := os.Getenv("GOOO_NAMED_INPUT_OUTPUT"); output != "" {
		raw, err := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": suite, "result": result, "flow": flow}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "fill.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("named assembly selection; separate native2/2; original selected fragment located through the named helper call; model0")
}
