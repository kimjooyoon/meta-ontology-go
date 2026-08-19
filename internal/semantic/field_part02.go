package semantic

import (
	"fmt"
)

func (f Field) Normalized() (Field, error) {
	id, err := ParseIdentity(f.ID.String())
	if err != nil {
		return Field{}, fmt.Errorf("%w: id: %v", ErrInvalidField, err)
	}
	parent, err := ParseIdentity(f.Parent.String())
	if err != nil {
		return Field{}, fmt.Errorf("%w: parent: %v", ErrInvalidField, err)
	}
	name, err := normalizeName(f.Name)
	if err != nil {
		return Field{}, fmt.Errorf("%w: name: %v", ErrInvalidField, err)
	}
	aliases, err := normalizeAliases(f.Aliases, name)
	if err != nil {
		return Field{}, fmt.Errorf("%w: aliases: %v", ErrInvalidField, err)
	}
	typeRef, err := f.TypeRef.Normalized()
	if err != nil {
		return Field{}, fmt.Errorf("%w: type ref: %v", ErrInvalidField, err)
	}
	if !f.Presence.Valid() {
		return Field{}, fmt.Errorf("%w: unknown presence %q", ErrInvalidField, f.Presence)
	}
	if !f.Cardinality.Valid() {
		return Field{}, fmt.Errorf("%w: unknown cardinality %q", ErrInvalidField, f.Cardinality)
	}
	span := f.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Field{}, fmt.Errorf("%w: span: %v", ErrInvalidField, err)
	}

	f.ID = id
	f.Parent = parent
	f.Name = name
	f.Aliases = aliases
	f.TypeRef = typeRef
	f.Span = span
	return f, nil
}
func (f Field) Validate() error {
	_, err := f.Normalized()
	return err
}
func (f Field) HasName(name string) bool {
	canonical, err := normalizeName(name)
	if err != nil {
		return false
	}
	if f.Name == canonical {
		return true
	}
	for _, alias := range f.Aliases {
		if alias == canonical {
			return true
		}
	}
	return false
}
