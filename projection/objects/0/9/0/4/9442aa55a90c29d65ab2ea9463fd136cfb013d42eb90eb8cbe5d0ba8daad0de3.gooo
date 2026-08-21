package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateFromProjectionV1UsesPackageOverrideInDigest(t *testing.T) {
	input := semanticIRProviderFixture{ir: acceptanceFixture()}
	before := copyIR(input.ir)
	result, err := GenerateFromProjectionV1(input, Options{PackageName: "adaptergen"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.SemanticIRDigest == "" || result.Metadata.SourceDigest == "" || result.Metadata.SourceMapDigest == "" {
		t.Fatalf("missing adapter projection digests: %#v", result.Metadata)
	}
	if !strings.Contains(string(result.Source), "package adaptergen") {
		t.Fatalf("package override was not applied:\n%s", result.Source)
	}
	if input.ir.Package != before.Package || !reflect.DeepEqual(input.ir, before) {
		t.Fatal("package override mutated caller-owned typed input")
	}
}
func TestGenerateFromProjectionV1HonorsHeaderOption(t *testing.T) {
	input := semanticIRProviderFixture{ir: acceptanceFixture()}
	before := copyIR(input.ir)
	result, err := GenerateFromProjectionV1(input, Options{Header: "// adapter projection header"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(result.Source, []byte("// adapter projection header\npackage ")) {
		t.Fatalf("projection header option was ignored:\n%s", result.Source)
	}
	if _, err := result.CanonicalJSON(); err != nil {
		t.Fatalf("header projection lost metadata binding: %v", err)
	}
	if !reflect.DeepEqual(input.ir, before) {
		t.Fatal("header projection mutated caller-owned typed input")
	}
}
func TestGenerateFromProjectionV1RejectsInvalidPackageWithoutMutation(t *testing.T) {
	input := semanticIRProviderFixture{ir: acceptanceFixture()}
	before := copyIR(input.ir)
	result, err := GenerateFromProjectionV1(input, Options{PackageName: "not-a-package", Header: "// ignored"})
	if err == nil || result.Metadata.SourceDigest != "" {
		t.Fatalf("invalid package returned projection metadata: result=%#v err=%v", result, err)
	}
	if !reflect.DeepEqual(input.ir, before) {
		t.Fatal("invalid package rejection mutated caller-owned typed input")
	}
}
func TestGeneratedSourceCompilesForNonStructOutput(t *testing.T) {
	ir := SemanticIR{
		Package: "compilegen",
		Activities: []Activity{{
			ID: "activity:render", Name: "Render", GoName: "Render",
			Outputs: []Port{{Name: "result", GoName: "result", GoType: "string"}},
		}},
	}
	result, err := Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result.Source), "return *new(string)") {
		t.Fatalf("non-struct output did not receive a compilable zero value:\n%s", result.Source)
	}
}
