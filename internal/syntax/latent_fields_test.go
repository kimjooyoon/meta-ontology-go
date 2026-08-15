package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsedEntitiesHaveNoReachableFields(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order`
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	if entity.Fields != nil {
		t.Fatalf("parser populated latent fields: %#v", entity.Fields)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatalf("existing source became unformattable: %v", err)
	}
	if !strings.Contains(formatted, `entity Order id "billing://entity/order"`) {
		t.Fatalf("formatted source lost entity declaration: %q", formatted)
	}
}

func TestProposedFieldSourceRemainsRejectedWithoutPartialFieldAST(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order" field Name id "billing://field/name" type string required one`
	first, firstDiagnostics := ParseFile("latent-fields.gooo", source)
	second, secondDiagnostics := ParseFile("latent-fields.gooo", source)
	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("proposed field source was not rejected deterministically")
	}
	if len(firstDiagnostics) == 0 || firstDiagnostics[0].Code != DiagUnexpectedDeclaration {
		t.Fatalf("field source diagnostics = %#v", firstDiagnostics)
	}
	entity := first.Declarations[0].(*EntityDecl)
	if len(entity.Fields) != 0 {
		t.Fatalf("rejected source produced partial field AST: %#v", entity.Fields)
	}
}

func TestSyntheticFieldsPreserveCarrierShapeAndCloneIsolation(t *testing.T) {
	first := FieldDecl{
		Span: Span{Filename: "synthetic.gooo", Start: Position{Offset: 10}, End: Position{Offset: 30}},
		ID:   "billing://field/first",
		Name: "First",
		TypeRef: TypeRefDecl{
			Span:     Span{Filename: "synthetic.gooo", Start: Position{Offset: 20}, End: Position{Offset: 26}},
			Spelling: "string",
		},
		Presence:        FieldPresenceRequired,
		Cardinality:     FieldCardinalityOne,
		IDSpan:          Span{Filename: "synthetic.gooo", Start: Position{Offset: 10}, End: Position{Offset: 18}},
		NameSpan:        Span{Filename: "synthetic.gooo", Start: Position{Offset: 19}, End: Position{Offset: 24}},
		PresenceSpan:    Span{Filename: "synthetic.gooo", Start: Position{Offset: 27}, End: Position{Offset: 35}},
		CardinalitySpan: Span{Filename: "synthetic.gooo", Start: Position{Offset: 36}, End: Position{Offset: 39}},
	}
	second := FieldDecl{
		Span:            Span{Filename: "synthetic.gooo", Start: Position{Offset: 40}, End: Position{Offset: 60}},
		ID:              "billing://field/second",
		Name:            "First",
		TypeRef:         TypeRefDecl{Span: Span{Filename: "synthetic.gooo", Start: Position{Offset: 50}, End: Position{Offset: 57}}, Spelling: "gooo:string"},
		Presence:        FieldPresenceOptional,
		Cardinality:     FieldCardinalityMany,
		IDSpan:          Span{Filename: "synthetic.gooo", Start: Position{Offset: 40}, End: Position{Offset: 48}},
		NameSpan:        Span{Filename: "synthetic.gooo", Start: Position{Offset: 49}, End: Position{Offset: 55}},
		PresenceSpan:    Span{Filename: "synthetic.gooo", Start: Position{Offset: 58}, End: Position{Offset: 66}},
		CardinalitySpan: Span{Filename: "synthetic.gooo", Start: Position{Offset: 67}, End: Position{Offset: 71}},
	}
	file := &File{
		Package:   &PackageDecl{Name: "billing"},
		Namespace: &NamespaceDecl{Name: "billing"},
		Decls: []Declaration{&EntityDecl{
			Name:   "Order",
			ID:     "billing://entity/order",
			Fields: []FieldDecl{first, second},
		}},
	}
	file.Declarations = file.Decls

	clone := file.Clone()
	if clone == file || clone.Decls[0] == file.Decls[0] {
		t.Fatal("file clone retained declaration aliases")
	}
	entity := clone.Decls[0].(*EntityDecl)
	if len(entity.Fields) != 2 || entity.Fields[0].ID != first.ID || entity.Fields[1].ID != second.ID {
		t.Fatalf("field order or IDs changed: %#v", entity.Fields)
	}
	if entity.Fields[0].Name != entity.Fields[1].Name {
		t.Fatalf("synthetic duplicate-looking fields were changed: %#v", entity.Fields)
	}
	if entity.Fields[0].TypeRef.Spelling != "string" || entity.Fields[1].TypeRef.Spelling != "gooo:string" {
		t.Fatalf("type reference spellings changed: %#v", entity.Fields)
	}
	if entity.Fields[0].Presence != FieldPresenceRequired || entity.Fields[1].Presence != FieldPresenceOptional || entity.Fields[0].Cardinality != FieldCardinalityOne || entity.Fields[1].Cardinality != FieldCardinalityMany {
		t.Fatalf("presence/cardinality changed: %#v", entity.Fields)
	}
	if entity.Fields[0].Span != first.Span || entity.Fields[0].TypeRef.Span != first.TypeRef.Span || entity.Fields[1].CardinalitySpan != second.CardinalitySpan {
		t.Fatalf("field spans changed: %#v", entity.Fields)
	}
	entity.Fields[0].ID = "changed"
	if file.Decls[0].(*EntityDecl).Fields[0].ID != first.ID {
		t.Fatal("field clone aliases caller-owned storage")
	}
}

func TestSyntheticFieldsAreNotSilentlyDroppedByFormatter(t *testing.T) {
	file := &File{
		Package:   &PackageDecl{Name: "billing"},
		Namespace: &NamespaceDecl{Name: "billing"},
		Decls: []Declaration{&EntityDecl{
			Name: "Order",
			ID:   "billing://entity/order",
			Fields: []FieldDecl{{
				ID:          "billing://field/name",
				Name:        "Name",
				TypeRef:     TypeRefDecl{Spelling: "string"},
				Presence:    FieldPresenceRequired,
				Cardinality: FieldCardinalityOne,
			}},
		}},
	}
	file.Declarations = file.Decls
	formatted, err := Format(file)
	if err != ErrLatentFieldsUnsupported || formatted != "" {
		t.Fatalf("latent field format result = %q, %v", formatted, err)
	}
}

func TestFieldEnumsExposeOnlyTechnologyIndependentValues(t *testing.T) {
	if !FieldPresenceRequired.Valid() || !FieldPresenceOptional.Valid() || FieldPresence("other").Valid() {
		t.Fatal("presence carrier values are not stable")
	}
	if !FieldCardinalityOne.Valid() || !FieldCardinalityMany.Valid() || FieldCardinality("many-items").Valid() {
		t.Fatal("cardinality carrier values are not stable")
	}
}
