package receiptprojection

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func projectReceiptField(field syntax.FieldDecl, byName map[string]*syntax.EntityDecl,
	ids, names, goNames map[string]bool) (Field, error) {
	// This independent receipt profile has no value-bound semantics. Refuse a
	// newer source constraint instead of projecting it away as an ordinary tag.
	if field.TextBound.IsSet() {
		return Field{}, fmt.Errorf("receipt profile %s does not support max_utf8_bytes on field %q", Profile, field.ID)
	}
	if err := claimID(ids, field.ID); err != nil {
		return Field{}, err
	}
	wireName := field.Name
	if wireName == strings.ToUpper(wireName) {
		wireName = strings.ToLower(wireName)
	}
	name := goFieldName(wireName)
	if !exportedName(name) || names[wireName] || goNames[name] {
		return Field{}, fmt.Errorf("invalid or colliding field name %q", field.Name)
	}
	names[wireName], goNames[name] = true, true
	target, ok := byName[field.TypeRef.Spelling]
	if !ok {
		return Field{}, fmt.Errorf("unresolved field type %q", field.TypeRef.Spelling)
	}
	goType, scalar := scalarTypes[target.ID]
	if !scalar {
		if !target.FieldsPresent || !exportedName(target.Name) {
			return Field{}, fmt.Errorf("unsupported type %q", target.Name)
		}
		goType = target.Name
	}
	if !field.Presence.Valid() || !field.Cardinality.Valid() {
		return Field{}, fmt.Errorf("unsupported presence or cardinality")
	}
	optional := field.Presence == syntax.FieldPresenceOptional
	many := field.Cardinality == syntax.FieldCardinalityMany
	if optional && many {
		return Field{}, fmt.Errorf("optional-many is outside receipt profile")
	}
	if !scalar && !optional && !many {
		return Field{}, fmt.Errorf("structural fields require optional-one or required-many; unbounded value recursion is rejected")
	}
	if many {
		goType = "[]" + goType
	} else if optional {
		goType = "*" + goType
	}
	return Field{ID: field.ID, Name: wireName, GoName: name, TypeID: target.ID,
		GoType: goType, Optional: optional, Many: many, Span: field.Span}, nil
}
