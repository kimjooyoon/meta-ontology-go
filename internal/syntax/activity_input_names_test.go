package syntax

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestAuthoredInputNamesSyntaxFirst(t *testing.T) {
	for _, tc := range []struct {
		signature string
		names     []string
	}{
		{"Integer", []string{"input"}},
		{"Integer, Integer", []string{"input0", "input1"}},
		{"minuend Integer, subtrahend Integer", []string{"minuend", "subtrahend"}},
		{"원금 Integer, 할인 Integer", []string{"원금", "할인"}},
		{"amount core.Integer, Integer", []string{"amount", "input1"}},
		{"Integer, adjustment Integer", []string{"input0", "adjustment"}},
		{"input0 Integer, named Integer", []string{"input0", "named"}},
	} {
		source := "package names\nnamespace names\n# 입력 이름과 타입의 위치를 각각 보관한다.\nactivity Difference(" + tc.signature + ") -> Integer id \"urn:names:difference\" computes `return 1`\n"
		f, diagnostics := ParseFile("names.gooo", source)
		if diagnostics.HasErrors() {
			t.Fatal(tc.signature, diagnostics)
		}
		a := f.Declarations[0].(*ActivityDecl)
		resolved, err := ActivityInputNames(a)
		if err != nil || !reflect.DeepEqual(resolved, tc.names) {
			t.Fatal(tc.signature, resolved, err)
		}
		for _, ref := range append(append([]NameRef{}, a.Inputs...), a.InputNames...) {
			if ref.Name != "" && source[ref.Span.Start.Offset:ref.Span.End.Offset] != ref.Name {
				t.Fatal("authored name/type source span changed", ref)
			}
		}
		formatted, err := Format(f)
		if err != nil || !strings.Contains(formatted, "Difference("+tc.signature+")") {
			t.Fatal(formatted, err)
		}
		roundtrip, diagnostics := ParseFile("formatted.gooo", formatted)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		again, err := ActivityInputNames(roundtrip.Declarations[0].(*ActivityDecl))
		if err != nil || !reflect.DeepEqual(again, resolved) {
			t.Fatal("name order changed", again, err)
		}
		if len(a.InputNames) > 0 {
			copy := f.Clone().Declarations[0].(*ActivityDecl)
			copy.InputNames[0].Name = "changed"
			if a.InputNames[0].Name == "changed" {
				t.Fatal("input name slice was shared")
			}
		} else {
			raw, _ := json.Marshal(a)
			if strings.Contains(string(raw), "InputNames") {
				t.Fatal("unnamed source acquired a new AST field")
			}
		}
	}
}

func TestAuthoredInputNamesRejectAmbiguousBindingsFirst(t *testing.T) {
	for _, signature := range []string{
		"same Integer, same Integer", "input1 Integer, other Integer", "other Integer, input0 Integer",
		"_ Integer", "nil Integer", "true Integer", "false Integer", "func Integer",
		"a core.", "a Integer b Integer",
	} {
		_, diagnostics := ParseFile("ambiguous.gooo", "package names\nnamespace names\nactivity Run("+signature+") -> Integer\n")
		if !diagnostics.HasErrors() {
			t.Fatal("ambiguous binding accepted", signature)
		}
	}
	file, diagnostics := ParseFile("manual.gooo", "package names\nnamespace names\nactivity Run(Integer) -> Integer\n")
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	a := file.Declarations[0].(*ActivityDecl)
	a.InputNames = []NameRef{{Name: "left"}, {Name: "right"}}
	if _, err := Format(file); err == nil {
		t.Fatal("misaligned input names formatted")
	}
}
