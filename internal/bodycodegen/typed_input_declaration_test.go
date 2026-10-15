package bodycodegen

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const inputDeclarationSourceFixture = `package inputpositions
namespace inputpositions
entity Integer id "urn:gooo:type:integer"
entity Pair id "urn:inputpositions:pair" fields {
 field left id "urn:inputpositions:left" type integer required one
 field right id "urn:inputpositions:right" type integer required one
}
# 동일한 타입의 두 입력을 선언 위치로 구분한다.
activity Difference(
 Integer,
 Integer
) -> Integer id "urn:inputpositions:difference" computes ` + "`return input0 - input1`" + `
activity Read(Pair) -> Integer id "urn:inputpositions:read" computes ` + "`return input.left + input.right`" + `
`

func TestInputDeclarationRepeatedTypesAndRename(t *testing.T) {
	var previous []TypedValueInput
	var previousSHA string
	for _, name := range []string{"Difference", "Subtract"} {
		source := []byte(strings.Replace(inputDeclarationSourceFixture, "Difference", name, 1))
		r, err := InspectTypedValueOrigins("입력.gooo", source, name)
		if err != nil || r.Status != "RESOLVED" || len(r.Inputs) != 2 {
			t.Fatal("two typed inputs", err, r.Reason)
		}
		for i, input := range r.Inputs {
			d := input.Declaration
			if d == nil || d.Status != "MAPPED" || d.View != "activity_signature" || d.InputIndex != i ||
				d.ActivityID != "urn:inputpositions:difference" || d.Port != input.Port ||
				d.TypeName != "Integer" || d.TypeID != "urn:gooo:type:integer" || d.SourceSHA256 != digest(source) ||
				string(source[d.Start:d.End]) != "Integer" {
				t.Fatal("signature identity", i, d)
			}
			if r.Sources[input.Node-1].Reason != "INPUT_PORT_IDENTITY" {
				t.Fatal("signature location was substituted for a body expression")
			}
		}
		if r.Inputs[0].Declaration.Start == r.Inputs[1].Declaration.Start {
			t.Fatal("same-type input positions collapsed")
		}
		if previous != nil {
			if previousSHA == r.SourceSHA256 || previous[0].Declaration.Start == r.Inputs[0].Declaration.Start ||
				previous[0].Declaration.ActivityID != r.Inputs[0].Declaration.ActivityID {
				t.Fatal("display rename lost stable activity identity or current source coordinates")
			}
		}
		previous, previousSHA = r.Inputs, r.SourceSHA256
	}
}

func TestInputDeclarationRecordFieldsKeepOneSignatureSlot(t *testing.T) {
	source := []byte(inputDeclarationSourceFixture)
	r, err := InspectTypedValueOrigins("record-input.gooo", source, "Read")
	if err != nil || r.Status != "RESOLVED" || len(r.Inputs) != 2 {
		t.Fatal("record input fields", err, r.Reason)
	}
	a, b := r.Inputs[0], r.Inputs[1]
	if a.FieldID == b.FieldID || a.Declaration == nil || b.Declaration == nil ||
		*a.Declaration != *b.Declaration || a.Declaration.TypeID != "urn:inputpositions:pair" || a.Declaration.Port != "input" {
		t.Fatal("field origin or shared signature slot lost", a, b)
	}
	before, err := TypedSourceControlInputs(r, "input/reader:Inspect")
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.Inputs {
		r.Inputs[i].Declaration = nil
	}
	after, err := TypedSourceControlInputs(r, "input/reader:Inspect")
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("additional signature facts changed the existing control wire", err)
	}
}

func TestInputDeclarationTextBindsExactSource(t *testing.T) {
	source := []byte(inputDeclarationSourceFixture)
	r, err := InspectTypedValueOrigins("입력.gooo", source, "Difference")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	before, _ := json.Marshal(r)
	var output bytes.Buffer
	if err := WriteTypedOriginText(&output, r, source); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`signature input0 at 10:2 "Integer"`, `signature input1 at 11:2 "Integer"`, `type="urn:gooo:type:integer"`} {
		if !strings.Contains(output.String(), want) {
			t.Fatal("readable input source missing", want, output.String())
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("renderer changed the source graph")
	}
	r.Inputs[0].Declaration.SourceSHA256 = "sha256:other-document"
	if err := WriteTypedOriginText(&bytes.Buffer{}, r, source); err == nil {
		t.Fatal("another document's signature accepted")
	}
}
