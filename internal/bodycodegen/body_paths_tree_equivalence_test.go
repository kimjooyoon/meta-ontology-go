package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTypedTreeEquivalencePreservesStructure(t *testing.T) {
	for _, tc := range []struct {
		name, original, candidate string
		equal                     bool
	}{
		{"parentheses", "return input - (2 - 3)", "return (input - ((2 - 3)))", true},
		{"literal spelling", "return input + 0x10", "return (input + 16)", true},
		{"constant", "return 2 - 3", "return (2 - 3)", true},
		{"constant operands", "return 2 - 3", "return (3 - 2)", false},
		{"constant literal", "return 0x10", "return 16", true},
		{"grouping", "return input - (2 - 3)", "return ((input - 2) - 3)", false},
		{"operands", "return input - 2", "return (2 - input)", false},
		{"name", "var x = input; return x", "var y = input; return y", false},
		{"assign target", "var x = input; var y = input; x = 3; return x+y", "var x = input; var y = input; y = 3; return x+y", false},
		{"statement order", "var x = input; x = x+1; x = x*2; return x", "var x = input; x = x*2; x = x+1; return x", false},
		{"branch", "if input < 0 { return input-1 } else { return input+1 }", "if input < 0 { return input+1 } else { return input-1 }", false},
		{"boolean operator", "if input < 0 && input < 2 { return input } else { return 0 }", "if input < 0 || input < 2 { return input } else { return 0 }", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generated := []byte("package sample\nfunc Assemble(input int64) int64 {" + tc.candidate + "}")
			receipt, err := typedBodyTreeEquivalence(context.Background(), "Assemble", tc.original, generated)
			if err != nil || receipt.Equivalent != tc.equal {
				t.Fatal("structure comparison", receipt, err)
			}
		})
	}
}

func TestSourceRecipeResourceBoundsAndScope(t *testing.T) {
	valid := recipeBytes(recipeOperand)
	source := recipeSource("return input-2")
	for _, raw := range [][]byte{nil, []byte("null"), append(append([]byte{}, valid...), 0xff), []byte(strings.Repeat(" ", 128<<10+1))} {
		if _, err := DecodeSourcePathDocument(context.Background(), "s.gooo", source, "Assemble", raw); err == nil {
			t.Fatal("unbounded or malformed document accepted")
		}
	}
	if _, err := DecodeSourcePathDocument(nil, "s.gooo", source, "Assemble", valid); err == nil {
		t.Fatal("nil context accepted")
	}
	for _, source := range [][]byte{nil, []byte(strings.Repeat(" ", 128<<10+1)), append(append([]byte{}, source...), 0xff)} {
		if _, err := DecodeSourcePathDocument(context.Background(), "s.gooo", source, "Assemble", valid); err == nil {
			t.Fatal("unbounded source accepted")
		}
	}
	var statements strings.Builder
	statements.WriteString("let x = input; ")
	for range 128 {
		statements.WriteString("x = input; ")
	}
	statements.WriteString("return x-2")
	for _, body := range []string{statements.String(), "return " + strings.Repeat("input + (", 20) + "2" + strings.Repeat(")", 20)} {
		if _, err := DecodeSourcePathDocument(context.Background(), "s.gooo", recipeSource(body), "Assemble", valid); err == nil {
			t.Fatal("arena bounds ignored")
		}
	}
	for _, choice := range []string{
		`{"id":"x","kind":"local_reference","occurrence":0,"intent":"choose","alternative_name":"later"}`,
		`{"id":"x","kind":"local_reference","occurrence":0,"intent":"choose","alternative_name":"flag"}`,
		`{"id":"x","kind":"root_order","occurrence":3,"intent":"choose"}`,
	} {
		body := "let first = input; let flag = true; let chosen = first; let later = input+1; if flag { return chosen } else { return later }"
		if _, err := DecodeSourcePathDocument(context.Background(), "s.gooo", recipeSource(body), "Assemble", recipeBytes(choice)); err == nil {
			t.Fatal("invalid scope, type or root schedule accepted", choice)
		}
	}
	// The node cap applies before SDK preparation, including wide expression trees.
	var many strings.Builder
	for i := range 64 {
		fmt.Fprintf(&many, "let x%d = input + %d; ", i, i)
	}
	many.WriteString("return x63")
	if _, err := DecodeSourcePathDocument(context.Background(), "s.gooo", recipeSource(many.String()), "Assemble", valid); err == nil {
		t.Fatal("expression cap ignored")
	}
}
