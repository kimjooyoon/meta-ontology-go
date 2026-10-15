package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestDirectRecordReturns22741(t *testing.T) {
	const source = `package directreturn
namespace directreturn22741
entity Integer id "urn:gooo:type:integer"
entity Box id "urn:directreturn22741:box" fields {
 field value id "urn:directreturn22741:value" type integer required one
 field tag id "urn:directreturn22741:tag" type string required one
}
activity Make(Integer)->Box id "urn:directreturn22741:make" computes "return Box{value:input+22741,tag:\"helper\"}"
activity Pick(Integer)->Box id "urn:directreturn22741:pick" computes ` + "`" + `if input<0 {return Make(input)}
if input==0 {return Box{value:22741,tag:"inline"}}
let local=Box{value:input,tag:"local"}
return local` + "`" + "\n"
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "origin/directreturn22741", Activity: "Pick"},
		Packages: []packageruntime.PackageSpec{{Path: "origin/directreturn22741", Name: "directreturn", Sources: []packageruntime.Source{{Filename: "직접반환22741.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"origin/directreturn22741:Pick":-83},"expected":{"origin/directreturn22741:Pick":{"value":22658,"tag":"helper"}}},
{"inputs":{"origin/directreturn22741:Pick":0},"expected":{"origin/directreturn22741:Pick":{"value":22741,"tag":"inline"}}},
{"inputs":{"origin/directreturn22741:Pick":97},"expected":{"origin/directreturn22741:Pick":{"value":97,"tag":"local"}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		file, saveErr := os.Create(filepath.Join(output, "direct-return22741.json.gz"))
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
	r := result.Runtime
	if r.FinitePassed != 3 || r.FiniteTotal != 3 || !r.RuntimeReplayed || r.ModelCalls != 0 || r.FieldObservation == nil {
		t.Fatal("functional and return-observation scope", r.FinitePassed, r.FieldObservation)
	}
	expressions := 0
	for _, binding := range r.FieldBindings {
		if binding.ExpressionKind == "return" {
			if binding.Declaration == nil || binding.Declaration.Status != "MAPPED" || binding.ExpressionOrdinal < 0 {
				t.Fatal("return source identity", binding)
			}
			expressions++
		}
	}
	if expressions != 3 || r.FieldObservation.RecordedWrites != 8 || r.FieldObservation.ParentLinks != 2 {
		t.Fatal("return binding/write/helper-parent populations", expressions, r.FieldObservation)
	}
	if r.FieldFileSources == nil {
		t.Fatal("original constructor sources absent")
	}
	mapped := 0
	for _, binding := range r.FieldFileSources.Bindings {
		if binding.Value.Status == "MAPPED" {
			mapped++
			if binding.Value.Filename != "직접반환22741.gooo" || !validEncodedReturnRange22741(source[binding.Value.Start:binding.Value.End], binding.Value.Expression) {
				t.Fatal("original constructor file range", binding)
			}
		}
	}
	if mapped < 6 {
		t.Fatal("original helper/inline/local constructor fields", mapped)
	}
	t.Log("new native3/3;helper/inline/named-local returns;three expression subjects,eight writes,two actual helper-return parent links;six constructor field anchors;two-run replay;model0")
}

func validEncodedReturnRange22741(encoded, decoded string) bool {
	if encoded == decoded {
		return true
	}
	value, err := strconv.Unquote("\"" + encoded + "\"")
	return err == nil && value == decoded
}
