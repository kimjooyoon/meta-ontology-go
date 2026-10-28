package bidir

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const recordReferenceSource = `package totals
namespace totals
entity Line id "urn:record-items:line" fields {
 field quantity id "urn:record-items:quantity" type integer required one
 field price id "urn:record-items:price" type integer required one
}
entity Basket id "urn:record-items:basket" fields {
 field items id "urn:record-items:items" type Line required many
}
`

func TestRecordReferencesPreserveSourceAndSemanticIdentity(t *testing.T) {
	support, ok := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	if !ok {
		t.Fatal("bounded record-reference source profile is unavailable")
	}
	for _, spelling := range []string{"Line", `"urn:record-items:line"`} {
		source := strings.Replace(recordReferenceSource, "type Line", "type "+spelling, 1)
		file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("records.gooo", source, support)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
		if err != nil {
			t.Fatal(err)
		}
		model, err := GetWithEntityFieldsSupport(document, support)
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range model.Nodes {
			if node.ID == "urn:record-items:basket" {
				if node.Fields[0].TypeRefUse.ResolvedID != "urn:record-items:line" || node.Fields[0].TypeRefUse.Spelling != strings.Trim(spelling, `"`) {
					t.Fatal(node.Fields[0])
				}
			}
		}
		if err := CheckGetPutWithEntityFieldsSupport(document, support); err != nil {
			t.Fatal(err)
		}
		if err := CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
			t.Fatal(err)
		}
		ir, err := LowerDocumentWithEntityFieldsSupport(document, support)
		if err != nil || ir.Validate() != nil {
			t.Fatal("record reference did not reach normalized IR", err)
		}
		if _, err := GetWithEntityFieldsSupport(document, EntityFieldsV6Support()); err == nil {
			t.Fatal("legacy profile accepted record references")
		}
	}
}

func TestRecordReferenceCyclesRetainTheClosingFieldOrigin(t *testing.T) {
	support, ok := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	if !ok {
		t.Fatal("bounded record-reference source profile is unavailable")
	}
	source := "package cycle\nnamespace cycle\n// 한글\nentity A id \"urn:cycle:a\" fields {\nfield b id \"urn:cycle:b-field\" type B required one\n}\n" +
		"entity B id \"urn:cycle:b\" fields {\nfield a id \"urn:cycle:a-field\" type A optional one\n}\n"
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("cycle.gooo", source, support)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	_, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	var located *EntityFieldsError
	if !errors.As(err, &located) || !strings.Contains(err.Error(), "cyclic record") || !located.Span.Valid() ||
		source[located.Span.Start:located.Span.End] != "A" {
		t.Fatal("cycle must identify its actual closing type reference", err)
	}
}

func TestRecordReferenceDepthAndDetachedRegistry(t *testing.T) {
	support, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	for _, depth := range []int{8, 9} {
		var source strings.Builder
		source.WriteString("package depth\nnamespace depth\n")
		for i := range depth {
			fmt.Fprintf(&source, "entity R%d id \"urn:depth:%d\" fields {\nfield next id \"urn:depth:%d:next\" type ", i, i, i)
			if i+1 < depth {
				fmt.Fprintf(&source, "R%d", i+1)
			} else {
				source.WriteString("integer")
			}
			source.WriteString(" required one\n}\n")
		}
		file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("depth.gooo", source.String(), support)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
		if depth == 9 {
			var located *EntityFieldsError
			if !errors.As(err, &located) || !strings.Contains(err.Error(), "depth exceeds 8") || !located.Span.Valid() {
				t.Fatal("excessive nested type depth accepted", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		registry := semantic.DefaultTypeRegistry()
		if _, err := getWithTypesAndEntityFieldsSupport(document, registry, support); err != nil {
			t.Fatal(err)
		}
		if _, err := registry.Resolve(semantic.TypeRef{Name: "R0"}); !errors.Is(err, semantic.ErrUnknownType) {
			t.Fatal("source compilation mutated the caller's type registry", err)
		}
		formatted, err := syntax.FormatWithEntityFieldsSupport(file, support)
		if err != nil {
			t.Fatal(err)
		}
		reparsed, diagnostics := syntax.ParseFileWithEntityFieldsSupport("formatted.gooo", formatted, support)
		if diagnostics.HasErrors() || len(reparsed.Declarations) != depth {
			t.Fatal("record layout did not survive formatting", diagnostics)
		}
	}
}
