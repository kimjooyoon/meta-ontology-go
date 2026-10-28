package syntaxregistration

import (
	"go/ast"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestRegistryAppendPreservesBoundedFamiliesAndCaseOrder(t *testing.T) {
	const path = "registry.go"
	raw := `package languagesyntax
func expectedRegistry() Registry {return Registry{Cases: all()}}
func all() []CaseDefinition {cases:=empty(); cases=append(cases,first()...); cases=append(cases,last()...); return cases}
func empty() []CaseDefinition {return []CaseDefinition{}}
func first() []CaseDefinition {return []CaseDefinition{{ID:"a"},{ID:"b"}}}
func last() []CaseDefinition {return []CaseDefinition{{ID:"c"},{ID:"d"}}}
`
	const decoy = "package languagesyntax\nfunc lastPart0() []CaseDefinition {return []CaseDefinition{{ID:\"decoy\"}}}\n"
	inputs := map[string][]byte{path: []byte(raw), "decoy.go": []byte(decoy)}
	for _, id := range []string{"e", "f", "g", "h", "i", "j"} {
		source, err := parseSourceUnits(inputs, "./", "languagesyntax", false)
		if err != nil {
			t.Fatal(err)
		}
		source.activity = "gooo://syntax-registration/generate-native"
		request := Request{Case: languagesyntax.CaseDefinition{ID: id, Path: "examples/" + id + ".gooo", ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax"}}
		if err := generateRegistry(source, request); err != nil {
			t.Fatal(err)
		}
		outputs, err := source.finish()
		if err != nil {
			t.Fatal(err)
		}
		if len(outputs) != 1 || outputs[path] == nil || string(inputs["decoy.go"]) != decoy {
			t.Fatal("native ownership or unrelated declaration changed")
		}
		inputs[path] = outputs[path]
	}
	observed, err := parseSourceUnits(inputs, "./", "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	root, err := observed.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	leaves, err := observed.registryCaseLeaves(root)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, leaf := range leaves {
		if len(leaf.Elts) > 2 {
			t.Fatal("existing source family width widened", len(leaf.Elts))
		}
		for _, element := range leaf.Elts {
			literal, ok := element.(*ast.CompositeLit)
			if !ok {
				t.Fatal("row origin lost")
			}
			id, err := registryLiteralField(literal, "ID")
			if err != nil {
				t.Fatal(err)
			}
			value, ok := id.(*ast.BasicLit)
			if !ok {
				t.Fatal("case identity lost")
			}
			decoded, err := strconv.Unquote(value.Value)
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, decoded)
		}
	}
	if strings.Join(ids, ",") != "a,b,c,d,e,f,g,h,i,j" {
		t.Fatal("case identities or order changed", ids)
	}
}

func TestRegistryAppendRefusesToDiscardComments(t *testing.T) {
	raw := `package languagesyntax
func expectedRegistry() Registry {return Registry{Cases: all()}}
func all() []CaseDefinition {cases:=first(); cases=append(cases,last()...); return cases}
func first() []CaseDefinition {return []CaseDefinition{{ID:"a"},{ID:"b"}}}
func last() []CaseDefinition {return []CaseDefinition{
// Preserve why this case exists.
{ID:"c"},{ID:"d"}}}
`
	source, err := parseSourceUnits(map[string][]byte{"registry.go": []byte(raw)}, "./", "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	err = generateRegistry(source, Request{Case: languagesyntax.CaseDefinition{ID: "e"}})
	if err == nil || !strings.Contains(err.Error(), "comment-preserving") || len(source.edits) != 0 {
		t.Fatal("commented source must remain intact when a preserving split is unavailable", err)
	}
}
