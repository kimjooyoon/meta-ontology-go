package bodycodegen

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

const namedInputOriginSource = `package namedorigins
namespace namedorigins
entity Integer id "urn:gooo:type:integer"
# 이름과 타입의 위치는 각각 원문에 연결된다.
activity Subtract(원금 Integer, 공제 Integer) -> Integer id "urn:namedorigins:subtract" computes "return 원금 - 공제"
activity Caller(amount Integer) -> Integer id "urn:namedorigins:caller" computes "return Subtract(amount, 17)"
`

func TestAuthoredInputNameOriginsFirst(t *testing.T) {
	source := []byte(namedInputOriginSource)
	r, err := InspectTypedValueOrigins("이름.gooo", source, "Subtract")
	if err != nil || r.Status != "RESOLVED" || len(r.Inputs) != 2 {
		t.Fatal(err, r.Reason)
	}
	for i, name := range []string{"원금", "공제"} {
		d := r.Inputs[i].Declaration
		if d == nil || d.Name != name || d.InputIndex != i || d.Port != []string{"input0", "input1"}[i] ||
			d.TypeID != "urn:gooo:type:integer" || string(source[d.NameStart:d.NameEnd]) != name ||
			string(source[d.Start:d.End]) != "Integer" || d.NameEnd >= d.Start {
			t.Fatal("named input lost positional/source identity", i, d)
		}
	}
	var text bytes.Buffer
	if err := WriteTypedOriginText(&text, r, source); err != nil || !strings.Contains(text.String(), `input name "원금" at 5:19`) {
		t.Fatal(err, text.String())
	}
	r.Inputs[0].Declaration.Name = "다른이름"
	if err := WriteTypedOriginText(&bytes.Buffer{}, r, source); err == nil {
		t.Fatal("name/source disagreement accepted")
	}
}

func TestAuthoredInputScopeAndRenameFirst(t *testing.T) {
	body := "let delta = 원금 - 공제; return delta"
	flow := InspectBodyValueOrigins(body, 2, "원금", "공제")
	inputs := map[string]bool{}
	for _, node := range flow.Nodes {
		if node.Kind == "input_reference" {
			inputs[node.Name] = true
		}
		if node.Kind == "unresolved_reference" {
			t.Fatal("named input unresolved", node)
		}
	}
	if flow.Status != "MAPPED" || len(inputs) != 2 || !inputs["원금"] || !inputs["공제"] {
		t.Fatal(flow)
	}
	refs, err := InspectNamedBodySourceReferences("return Subtract(1)", 1, []string{"Subtract"})
	if err != nil || len(refs) != 1 || refs[0].Direct {
		t.Fatal("input mistaken for a helper", refs, err)
	}
	r, err := RenameActivitySource(context.Background(), "names.gooo", []byte(namedInputOriginSource), "Subtract", "Difference")
	if err != nil || r.CallSites != 1 || !strings.Contains(r.Source, "Difference(amount, 17)") {
		t.Fatal(r, err)
	}
	if _, _, err := renameBodyCalls("return Subtract(amount, 17)", 1, "Subtract", "amount", "amount"); err == nil {
		t.Fatal("rename captured an authored input")
	}
}
