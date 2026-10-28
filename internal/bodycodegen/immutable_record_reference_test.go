package bodycodegen

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"strconv"
	"strings"
	"testing"
)

const immutableReferenceDeclarations = `package aliasboundary
namespace aliasboundary
entity Child id "urn:alias:child" fields {
 field amount id "urn:alias:amount" type integer required one
}
entity Box id "urn:alias:box" fields {
 field child id "urn:alias:child-field" type Child optional one
}
entity Result id "urn:alias:result" fields {
 field value id "urn:alias:value" type integer required one
}
`

func immutableReferenceProgram(body string) []byte {
	return []byte(immutableReferenceDeclarations + "activity Adjust(Box) -> Result id \"urn:alias:adjust\" computes " + strconv.Quote(body) + "\n")
}

func TestImmutableRecordReferencesRejectFieldWrites(t *testing.T) {
	input := `[{"child":{"amount":7}}]`
	for _, c := range []struct{ name, body string }{
		{"direct", "let child = input.child\nif child != nil { child.amount = 8 }\nreturn Result{value: 0}"},
		{"alias", "let original = input.child\nlet child = original\nif child != nil { child.amount = 8 }\nreturn Result{value: 0}"},
		{"compound", "let child = input.child\nif child != nil { child.amount += 1 }\nreturn Result{value: 0}"},
		{"parenthesized", "let child = input.child\nif child != nil { (child).amount = 8 }\nreturn Result{value: 0}"},
		{"callee", "return Mutate(input)"},
		{"range", "for row in input.items { let child = row.child; if child != nil { child.amount = 8 } }\nreturn Result{value: 0}"},
	} {
		t.Run(c.name, func(t *testing.T) {
			source := immutableReferenceProgram(c.body)
			observedInput := input
			if c.name == "callee" {
				source = append(source, []byte("activity Mutate(Box) -> Result id \"urn:alias:mutate\" computes "+strconv.Quote("let child = input.child\nif child != nil { child.amount = 8 }\nreturn Result{value: 0}")+"\n")...)
			}
			if c.name == "range" {
				text := strings.Replace(string(source), "entity Box id", `entity Wrapper id "urn:alias:wrapper" fields { field child id "urn:alias:wrapper-child" type Child optional one }
entity Box id`, 1)
				source = []byte(strings.Replace(text, " field child id \"urn:alias:child-field\" type Child optional one", " field child id \"urn:alias:child-field\" type Child optional one\n field items id \"urn:alias:items\" type Wrapper required many", 1))
				observedInput = `[{"items":[{"child":{"amount":7}}]}]`
			}
			generated, err := GenerateSourceBundle([]sourcebundle.Source{{Filename: "alias.gooo", Content: string(source)}}, "Adjust")
			if err == nil || !strings.Contains(err.Error(), "immutable record reference") || generated.Source != "" {
				t.Fatalf("native projection accepted a reference write: %v", err)
			}
			value, err := EvaluateFixedValue(context.Background(), "alias.gooo", source, "Adjust", []byte(observedInput))
			if err == nil || !strings.Contains(err.Error(), "immutable record reference") || len(value.Value) != 0 {
				t.Fatalf("fixed evaluation accepted a reference write: %v", err)
			}
		})
	}
}

func assertImmutableReferenceValues(t *testing.T, source []byte, cases [][2]string) {
	t.Helper()
	generated, err := GenerateSourceBundle([]sourcebundle.Source{{Filename: "values.gooo", Content: string(source)}}, "Adjust")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		value, err := EvaluateFixedValue(context.Background(), "values.gooo", source, "Adjust", []byte("["+c[0]+"]"))
		if err != nil || string(value.Value) != c[1] {
			t.Fatalf("fixed value: %s / %v", value.Value, err)
		}
	}
	assertNominalNativeCases(t, generated.Result, "Box", cases)
}

func TestNominalReferenceBoundaryPreservesRequiredRecordCopies(t *testing.T) {
	source := immutableReferenceProgram("let child = input.child\nchild.amount += 1\nreturn Result{value: child.amount}")
	source = []byte(strings.Replace(string(source), "type Child optional one", "type Child required one", 1))
	assertImmutableReferenceValues(t, source, [][2]string{
		{`{"child":{"amount":7}}`, `{"value":8}`},
		{`{"child":{"amount":9007199254740993}}`, `{"value":9007199254740994}`},
	})
}

func TestNominalReferenceBoundaryPreservesOptionalWholeFieldReplacement(t *testing.T) {
	source := immutableReferenceProgram("let result = input\nif input.child != nil { let child = Child{amount: input.child.amount + 1}; result.child = &child }\nreturn result")
	source = []byte(strings.Replace(string(source), "activity Adjust(Box) -> Result", "activity Adjust(Box) -> Box", 1))
	assertImmutableReferenceValues(t, source, [][2]string{
		{`{}`, `{}`},
		{`{"child":{"amount":7}}`, `{"child":{"amount":8}}`},
	})
}
