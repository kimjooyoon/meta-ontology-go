package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const namedRecordSource = `package namedrecord
namespace namedrecord
entity Integer id "urn:gooo:type:integer"
entity Amount id "urn:namedrecord:amount" fields {
 field value id "urn:namedrecord:value" type integer required one
}
activity Read(amount Amount) -> Integer id "urn:namedrecord:read" computes "return amount.value"
activity Scale(Amount, factor Integer) -> Integer id "urn:namedrecord:scale" computes "if factor < 0 { return 0 }; return Read(input0) * factor"
`

func TestAuthoredInputRecordAndMixedNativeFirst(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "named/record", Activity: "Scale"},
		Packages: []packageruntime.PackageSpec{{Path: "named/record", Name: "namedrecord",
			Sources: []packageruntime.Source{{Filename: "record.gooo", Content: namedRecordSource}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"named/record:Scale.input0":{"value":37},"named/record:Scale.input1":3},"expected":{"named/record:Scale":111}},
 {"inputs":{"named/record:Scale.input0":{"value":53},"named/record:Scale.input1":-2},"expected":{"named/record:Scale":0}}
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
	origins, err := bodycodegen.InspectTypedValueOrigins("record.gooo", []byte(namedRecordSource), "Read")
	if err != nil || origins.Status != "RESOLVED" || len(origins.Inputs) != 1 {
		t.Fatal(origins, err)
	}
	input := origins.Inputs[0]
	if input.Port != "input" || input.FieldID != "urn:namedrecord:value" || input.Declaration == nil ||
		input.Declaration.Name != "amount" || input.Declaration.TypeID != "urn:namedrecord:amount" {
		t.Fatal("field and input-name origins conflated", input)
	}
	if output := os.Getenv("GOOO_NAMED_INPUT_OUTPUT"); output != "" {
		raw, err := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": suite, "result": result, "origins": origins}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "record.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("2/2 output obligations; named record helper, mixed signature, branch and distinct field origin; model0")
}
