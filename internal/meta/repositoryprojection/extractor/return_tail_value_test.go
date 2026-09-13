package extractor

import (
	"go/ast"
	"testing"
)

const returnTailValueTypes = "package p\n" +
	"type Leaf struct { Label string }\n" +
	"type Report struct { Label string; Nested Leaf; Link *Leaf; Values []string }\n" +
	"type Embedded struct { Leaf }\n" +
	"type Promoted struct { *Leaf }\n"

func TestReturnTailReturnedValueFieldBoundary(t *testing.T) {
	cases := []struct {
		name     string
		function string
		allowed  bool
	}{
		{"returned-field", "func F(r Report) Report { r.Label = \"new\"; return r }", true},
		{"nested-value", "func F(r Report) Report { r.Nested.Label = \"new\"; return r }", true},
		{"embedded-value", "func F(r Embedded) Embedded { r.Label = \"new\"; return r }", true},
		{"parenthesized-value", "func F(r Report) Report { (r).Label = \"new\"; return (r) }", true},
		{"replaced-slice-header", "func F(r Report) Report { r.Values = []string{\"new\"}; return r }", true},
		{"pointer-root", "func F(r *Report) *Report { r.Label = \"new\"; return r }", false},
		{"explicit-dereference", "func F(r *Report) *Report { (*r).Label = \"new\"; return r }", false},
		{"pointer-field", "func F(r Report) Report { r.Link.Label = \"new\"; return r }", false},
		{"embedded-pointer", "func F(r Promoted) Promoted { r.Label = \"new\"; return r }", false},
		{"indexed-storage", "func F(r Report) Report { r.Values[0] = \"new\"; return r }", false},
		{"whole-value-rebinding", "func F(r Report) Report { r = Report{}; return r }", false},
		{"updated-value-not-returned", "func F(r Report) Report { r.Label = \"new\"; return Report{} }", false},
		{"scalar-rebinding", "func F(r int) int { r = 2; return r }", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			function, info := returnTailTypedFixture(t, returnTailValueTypes+tc.function)
			object := info.Defs[function.Type.Params.List[0].Names[0]]
			tail := function.Body.List[len(function.Body.List)-2:]
			assignment := tail[0].(*ast.AssignStmt)
			if got := returnTailReturnedValueField(assignment.Lhs[0], object, tail, info); got != tc.allowed {
				t.Fatalf("value-field permission=%t, want %t", got, tc.allowed)
			}
			err := hasReturnTailBindingHazard(function.Body, tail, []suffixBinding{{object: object}}, info)
			if (err == nil) != tc.allowed {
				t.Fatalf("binding proof error=%v, allowed=%t", err, tc.allowed)
			}
		})
	}
}

func TestReturnTailReturnedValueFieldEscapesRejected(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
	}{
		{"address-before-tail", "alias := &r; _ = alias;"},
		{"closure-before-tail", "_ = func() string { return r.Label };"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := returnTailValueTypes + "func F(r Report) Report { " + tc.prefix + " r.Label = \"new\"; return r }"
			function, info := returnTailTypedFixture(t, source)
			object := info.Defs[function.Type.Params.List[0].Names[0]]
			tail := function.Body.List[len(function.Body.List)-2:]
			if err := hasReturnTailBindingHazard(function.Body, tail, []suffixBinding{{object: object}}, info); err == nil {
				t.Fatal("escaped returned value was accepted")
			}
		})
	}
}

func TestReturnTailReturnedValueFieldMissingEvidenceRejected(t *testing.T) {
	for _, missing := range []string{"field-object", "receiver-type-and-object"} {
		t.Run(missing, func(t *testing.T) {
			function, info := returnTailTypedFixture(t, returnTailValueTypes+"func F(r Report) Report { r.Label = \"new\"; return r }")
			object := info.Defs[function.Type.Params.List[0].Names[0]]
			tail := function.Body.List
			selector := tail[0].(*ast.AssignStmt).Lhs[0].(*ast.SelectorExpr)
			if missing == "field-object" {
				delete(info.Uses, selector.Sel)
			} else {
				delete(info.Types, selector.X)
				delete(info.Uses, selector.X.(*ast.Ident))
			}
			if returnTailReturnedValueField(selector, object, tail, info) {
				t.Fatal("missing typed selection evidence was accepted")
			}
		})
	}
}
