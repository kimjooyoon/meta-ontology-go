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

func TestScopedCounterAssembly18717(t *testing.T) {
	const source = `package counterfill
namespace counterfill18717
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer)->Integer id "urn:counterfill18717:sum" computes "let total=input;for let i=0;i<6;i++ {total=total+i+__GOOO_BODY_HOLE_bump__};return total" assembling {
 source_fill intent "범위 안의 카운터와 선언된 증가량을 더한다" {
  hole "bump"
  candidate "nine" {fill "bump" "9"}
  candidate "seventeen" {fill "bump" "17"}
 }
 case "31" -> "148"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/counterfill18717", Activity: "Sum"}, Packages: []packageruntime.PackageSpec{{Path: "origin/counterfill18717", Name: "counterfill", Sources: []packageruntime.Source{{Filename: "counter18717.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/counterfill18717:Sum":149},"expected":{"origin/counterfill18717:Sum":266}},{"inputs":{"origin/counterfill18717:Sum":-151},"expected":{"origin/counterfill18717:Sum":-34}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal("new counted assembly", err, result.Runtime.FinitePassed)
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		f, e := os.Create(filepath.Join(root, "counter-fill18717.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if e = json.NewEncoder(z).Encode(result); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	selected := 0
	for _, step := range result.BodyFills {
		fill := step.Generation.Report.BodyFill
		if fill != nil && fill.SelectedCandidateID == "seventeen" && fill.TestCasesPassed == 1 && fill.TestCasesTotal == 1 && fill.Evaluator == "gooo/bodycodegen-int64-ast-interpreter/v2" {
			selected++
		}
	}
	if selected != 1 || result.Runtime.ModelCalls != 0 || !result.Runtime.RuntimeReplayed {
		t.Fatal("bounded selection/native scope", selected, result.Runtime.ModelCalls)
	}
	updates := 0
	for _, binding := range result.Runtime.LocalFileSources.Bindings {
		if binding.Value.Expression == "i++" {
			if binding.Value.Status != "MAPPED" || source[binding.Value.Start:binding.Value.End] != "i++" {
				t.Fatal("assembled increment original source", binding)
			}
			updates++
		}
	}
	if updates != 1 {
		t.Fatal("authored update sites", updates)
	}

	t.Log("new Gooo source-fill with for-let/++:typed interpreter chooses seventeen against1 declared case;separate native2/2;original increment location retained;two-run native replay;model0")
}
