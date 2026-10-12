package extractor

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestReturnTailFreshLiteralAddressBindings(t *testing.T) {
	cases := []struct {
		name, expression string
		allowed          bool
	}{
		{"fresh-record", "&struct{ Value int }{Value: value}", true},
		{"fresh-array", "&[1]int{value}", true},
		{"fresh-slice-header", "&[]int{value}", true},
		{"fresh-map-header", "&map[int]int{value: 1}", true},
		{"parenthesized-literal", "&(struct{ Value int }{Value: value})", true},
		{"existing-binding", "&value", false},
		{"nested-binding-address", "&struct{ Link *int }{Link: &value}", false},
	}
	for _, tc := range cases {
		for _, prefix := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/tail", true: "/prefix"}[prefix], func(t *testing.T) {
				function, info := returnTailTypedFixture(t, "package p\nfunc F(value int) int { _ = "+tc.expression+"; return value }")
				object := info.Defs[function.Type.Params.List[0].Names[0]]
				tail := function.Body.List
				if prefix {
					tail = tail[1:]
				}
				err := hasReturnTailBindingHazard(function.Body, tail, []suffixBinding{{object: object}}, info)
				if (err == nil) != tc.allowed {
					t.Fatalf("allowed=%t err=%v", tc.allowed, err)
				}
			})
		}
	}
}

func TestReturnTailFreshLiteralRequiresTypeEvidence(t *testing.T) {
	for _, missing := range []string{"literal", "address", "mismatch"} {
		t.Run(missing, func(t *testing.T) {
			function, info := returnTailTypedFixture(t, "package p\nfunc F(value int) int { _ = &struct{ Value int }{Value: value}; return value }")
			object := info.Defs[function.Type.Params.List[0].Names[0]]
			var address *ast.UnaryExpr
			ast.Inspect(function.Body, func(n ast.Node) bool {
				if x, ok := n.(*ast.UnaryExpr); ok && x.Op == token.AND {
					address = x
				}
				return true
			})
			switch missing {
			case "literal":
				delete(info.Types, address.X)
			case "address":
				delete(info.Types, address)
			case "mismatch":
				v := info.Types[address]
				v.Type = types.NewPointer(types.Typ[types.Int])
				info.Types[address] = v
			}
			if err := hasReturnTailBindingHazard(function.Body, function.Body.List, []suffixBinding{{object: object}}, info); err == nil {
				t.Fatal("missing or inconsistent address evidence accepted")
			}
		})
	}
}

func TestReturnTailFreshLiteralExtractionRetainsProof(t *testing.T) {
	source := "package p\nfunc F(value int) int {\n" + strings.Repeat(" _ = 1\n", 80) +
		" _ = &struct{ Value int }{Value: value}\n return value\n}\n"
	result, err := extractReturnTailRegressionFixture(t, source)
	if err != nil {
		t.Fatal(err)
	}
	assertReturnTailTupleProofChain(t, result)
	if !generatedContains(result.Generated, "&struct{ Value int }{Value: value}") {
		t.Fatal("fresh literal was not preserved")
	}
}

func TestReturnTailFreshLiteralDoesNotProveOpaqueCallee(t *testing.T) {
	source := "package p\nfunc F(value int) int {\n" + strings.Repeat(" _ = 1\n", 80) +
		" _ = &struct{ Value int }{Value: value}\n return opaque(value)\n}\n" +
		"func opaque(value int) int { panic(value) }\n"
	_, err := extractReturnTailRegressionFixture(t, source)
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "CALLEE_EFFECTS_UNPROVEN" {
		t.Fatalf("opaque call unexpectedly admitted: %v", err)
	}
	found := false
	for _, diagnostic := range failure.Diagnostics {
		found = found || strings.HasPrefix(diagnostic, "callee=") && strings.HasSuffix(diagnostic, ".opaque")
	}
	if !found {
		t.Fatalf("missing callee location: %+v", failure)
	}
}
