package bodycodegen

import "testing"

func TestBodySourceReferencesRespectLexicalInputs(t *testing.T) {
	for _, tc := range []struct {
		body   string
		arity  int
		name   string
		direct bool
	}{
		{"return Fee(input)", 1, "Fee", true},
		{"return (Fee)(input)", 1, "Fee", true},
		{"return body(input)", 1, "body", true},
		{"return input(input)", 1, "input", false},
		{"return input0(input1)", 2, "input0", false},
		{"let Fee = input; return Fee(input)", 1, "Fee", false},
		{"return external.Fee(input)", 1, "external.Fee", false},
	} {
		refs, err := InspectBodySourceReferences(tc.body, tc.arity)
		if err != nil || len(refs) != 1 || refs[0].Name != tc.name || refs[0].Direct != tc.direct || refs[0].Ordinal != 0 {
			t.Fatalf("%s: %+v %v", tc.body, refs, err)
		}
	}
}

func TestBodySourceReferencesKeepNestedCallsAndUnknownSyntax(t *testing.T) {
	refs, err := InspectBodySourceReferences("return Outer(Inner(input))", 1)
	if err != nil || len(refs) != 2 || refs[0].Name != "Outer" || refs[1].Name != "Inner" || refs[1].Ordinal != 1 {
		t.Fatal(refs, err)
	}
	for _, body := range []string{"return (", "let = input"} {
		if _, err := InspectBodySourceReferences(body, 1); err == nil {
			t.Fatalf("accepted malformed syntax: %q", body)
		}
	}
	if _, err := InspectBodySourceReferences("return input", 17); err == nil {
		t.Fatal("unbounded parameter list")
	}
}

func TestBodySourceFragmentRetainsUnresolvedEnclosingScope(t *testing.T) {
	refs, err := InspectBodySourceReferences("return Step(input)", 1, "let Step = input; return 0")
	if err != nil || len(refs) != 1 || refs[0].Direct || !refs[0].AmbiguousScope || refs[0].Name != "Step" {
		t.Fatal(refs, err)
	}
	// A declaration in another block also needs candidate location evidence.
	refs, err = InspectBodySourceReferences("return Step(input)", 1, "if input < 0 { let Step = input; _ = Step }; return 0")
	if err != nil || !refs[0].AmbiguousScope {
		t.Fatal(refs, err)
	}
	refs, err = InspectBodySourceReferences("return Step(input)", 1, "let Offset = input; return Offset")
	if err != nil || !refs[0].Direct || refs[0].AmbiguousScope {
		t.Fatal(refs, err)
	}
	if _, err := InspectBodySourceReferences("return Step(input)", 1, "return ("); err == nil {
		t.Fatal("invalid enclosing context silently discarded")
	}
}
