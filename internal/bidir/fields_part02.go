package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func (f Field) semantic() (semantic.Field, error) {
	id, err := semantic.ParseIdentity(string(f.ID))
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: id: %v", ErrInvalidField, err)
	}
	parent, err := semantic.ParseIdentity(string(f.Parent))
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: parent: %v", ErrInvalidField, err)
	}
	typeRef := f.TypeRef
	if typeRef.ID == "" && typeRef.Name == "" && f.TypeRefUse.ResolvedID != "" {
		typeRef = semantic.TypeRef{ID: semantic.ID(f.TypeRefUse.ResolvedID)}
	}
	if typeRef.ID == "" && typeRef.Name == "" && f.TypeRefUse.Spelling != "" {
		var typeErr error
		typeRef, typeErr = typeRefFromUse(f.TypeRefUse)
		if typeErr != nil {
			return semantic.Field{}, fmt.Errorf("%w: type reference use: %v", ErrInvalidField, typeErr)
		}
	}
	field, err := (semantic.Field{
		ID:          id,
		Parent:      parent,
		Name:        f.Name,
		Aliases:     append([]string(nil), f.Aliases...),
		TypeRef:     typeRef,
		Presence:    f.Presence,
		Cardinality: f.Cardinality,
		Span:        toSemanticSpan(f.Span),
	}).Normalized()
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: %v", ErrInvalidField, err)
	}
	return field, nil
}
