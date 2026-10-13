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

func TestWorkspaceFilledConditionOrigin3889(t *testing.T) {
	const source = `package order3889
namespace order3889
entity Integer id "urn:gooo:type:integer"
activity Price(Integer) -> Integer id "urn:order3889:price" computes "let adjustment = __GOOO_BODY_HOLE_adjustment__; if input > __GOOO_BODY_HOLE_threshold__ { return input + adjustment }; return input - 13" assembling {
 source_fill intent "조건 일부와 금액을 함께 조립한다" {
  hole "adjustment"
  hole "threshold"
  candidate "bounded3889" { fill "adjustment" "43" fill "threshold" "73" }
  candidate "other3889" { fill "adjustment" "2" fill "threshold" "11" }
 }
 case "79" -> "122"
 case "17" -> "4"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "shop3889", Activity: "Price"},
		Packages: []packageruntime.PackageSpec{{Path: "shop3889", Name: "order3889",
			Sources: []packageruntime.Source{{Filename: "price3889.gooo", Content: source}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"shop3889:Price":89},"expected":{"shop3889:Price":132}},{"inputs":{"shop3889:Price":31},"expected":{"shop3889:Price":18}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || len(result.BodyFills) != 1 || len(result.Runtime.ConditionSites) != 1 {
		t.Fatal("workspace source-fill origin", err, result.Runtime.FinitePassed)
	}
	link := result.Runtime.ConditionSites[0].Source.Construction
	if link == nil || link.CandidateID != "bounded3889" || len(link.Fragments) != 2 ||
		link.SourceSHA256 != result.BodyFills[0].InputSourceSHA256 {
		t.Fatal("package pre-fill did not reach final condition", link)
	}
	checkWorkspaceOrigin3889(t, result)
	if _, _, err := reconstructWorkspace(context.Background(), manifest, result); err != nil {
		t.Fatal("fresh saved construction cannot retain its origin", err)
	}
	if result.Runtime.ModelCalls != 0 {
		t.Fatal("unexpected model inference")
	}
	if output := os.Getenv("GOOO_FILL_ORIGIN_OUTPUT"); output != "" {
		raw, err := json.MarshalIndent(map[string]any{"manifest": manifest, "result": result}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "workspace3889.json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("workspace3889: native2/2; pre-filled threshold -> final condition, retained template and selected candidate fragments, assembly input offsets; fresh saved construction read; model0")
}

func checkWorkspaceOrigin3889(t *testing.T, result Result) {
	t.Helper()
	link := result.Runtime.ConditionSites[0].Source.Construction
	counts := map[string]int{}
	for _, fragment := range link.Fragments {
		counts[fragment.Kind]++
		if fragment.Kind == "declared_candidate" {
			if fragment.HoleID != "threshold" || fragment.ExpressionSource == nil ||
				result.Program.Source[fragment.ExpressionSource.Start:fragment.ExpressionSource.End] != "73" {
				t.Fatal("package candidate origin range differs", fragment)
			}
		}
	}
	if counts["declared_candidate"] != 1 || counts["retained_template"] != 1 {
		t.Fatal("partial condition origin fragments", counts)
	}
	if len(bodyexecution.ConditionFillOriginFacts(result.Runtime)) != 2 {
		t.Fatal("package origin facts unavailable")
	}
}
