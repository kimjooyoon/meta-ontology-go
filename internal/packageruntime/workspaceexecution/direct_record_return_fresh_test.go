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

func TestDirectRecordReturns21917(t *testing.T) {
	const source = `package directreturn
namespace directreturn21917
entity Integer id "urn:gooo:type:integer"
entity Box id "urn:directreturn21917:box" fields {
 field value id "urn:directreturn21917:value" type integer required one
 field tag id "urn:directreturn21917:tag" type string required one
}
activity Make(Integer)->Box id "urn:directreturn21917:make" computes "return Box{value:input+21917,tag:\"helper\"}"
activity Pick(Integer)->Box id "urn:directreturn21917:pick" computes ` + "`" + `if input<0 {return Make(input)}
if input==0 {return Box{value:21917,tag:"inline"}}
let local=Box{value:input,tag:"local"}
return local` + "`" + "\n"
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "origin/directreturn21917", Activity: "Pick"},
		Packages: []packageruntime.PackageSpec{{Path: "origin/directreturn21917", Name: "directreturn", Sources: []packageruntime.Source{{Filename: "직접반환21917.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"origin/directreturn21917:Pick":-59},"expected":{"origin/directreturn21917:Pick":{"value":21858,"tag":"helper"}}},
{"inputs":{"origin/directreturn21917:Pick":0},"expected":{"origin/directreturn21917:Pick":{"value":21917,"tag":"inline"}}},
{"inputs":{"origin/directreturn21917:Pick":67},"expected":{"origin/directreturn21917:Pick":{"value":67,"tag":"local"}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		file, saveErr := os.Create(filepath.Join(output, "direct-return21917.json.gz"))
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
			if binding.Value.Filename != "직접반환21917.gooo" || !validEncodedReturnRange21917(source[binding.Value.Start:binding.Value.End], binding.Value.Expression) {
				t.Fatal("original constructor file range", binding)
			}
		}
	}
	if mapped < 6 {
		t.Fatal("original helper/inline/local constructor fields", mapped)
	}
	t.Log("new native3/3;helper/inline/named-local returns;three expression subjects,eight writes,two actual helper-return parent links;six constructor field anchors;two-run replay;model0")
}

func validEncodedReturnRange21917(encoded, decoded string) bool {
	if encoded == decoded {
		return true
	}
	value, err := strconv.Unquote("\"" + encoded + "\"")
	return err == nil && value == decoded
}
