package syntax

import (
	"testing"
)

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
