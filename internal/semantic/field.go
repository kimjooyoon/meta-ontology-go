package semantic

import (
	"errors"
	"fmt"
)

var ErrInvalidField = errors.New("invalid semantic field")

type Presence string

const (
	Required Presence = "required"
	Optional Presence = "optional"

	FieldRequired = Required
	FieldOptional = Optional
)

func (p Presence) Valid() bool {
	return p == Required || p == Optional
}

type Cardinality string

const (
	One  Cardinality = "one"
	Many Cardinality = "many"

	CardinalityOne  = One
	CardinalityMany = Many
)

func (c Cardinality) Valid() bool {
	return c == One || c == Many
}

type FieldPresence = Presence
type FieldCardinality = Cardinality

// Field is an ordered structural member of exactly one Entity. Name and
// Aliases are presentation metadata; ID and Parent are explicit identity
// boundaries and never derive from names or paths.
type Field struct {
	ID          ID
	Parent      ID
	Name        string
	Aliases     []string
	TypeRef     TypeRef
	Presence    Presence
	Cardinality Cardinality
	Span        Span
}

func NewField(id, parent ID, name string, typeRef TypeRef, presence Presence, cardinality Cardinality) (Field, error) {
	return (Field{
		ID:          id,
		Parent:      parent,
		Name:        name,
		TypeRef:     typeRef,
		Presence:    presence,
		Cardinality: cardinality,
	}).Normalized()
}

func NewFieldFromStrings(id, parent, name string, typeID ID, presence Presence, cardinality Cardinality) (Field, error) {
	parsedID, err := ParseIdentity(id)
	if err != nil {
		return Field{}, err
	}
	parsedParent, err := ParseIdentity(parent)
	if err != nil {
		return Field{}, err
	}
	return NewField(parsedID, parsedParent, name, TypeRef{ID: typeID}, presence, cardinality)
}

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

func normalizeFields(raw []Field, parent ID, kind Kind) ([]Field, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if kind != Entity {
		return nil, fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
	}
	fields := make([]Field, 0, len(raw))
	fieldIDs := make(map[ID]struct{}, len(raw))
	nameOwners := make(map[string]ID, len(raw)*2)
	for _, field := range raw {
		normalized, err := field.Normalized()
		if err != nil {
			return nil, err
		}
		if normalized.Parent != parent {
			return nil, fmt.Errorf("%w: field %s parent is %s, want %s", ErrInvalidField, normalized.ID, normalized.Parent, parent)
		}
		if _, exists := fieldIDs[normalized.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate field ID %s", ErrInvalidField, normalized.ID)
		}
		fieldIDs[normalized.ID] = struct{}{}
		for _, name := range append([]string{normalized.Name}, normalized.Aliases...) {
			if owner, exists := nameOwners[name]; exists && owner != normalized.ID {
				return nil, fmt.Errorf("%w: field name %q is shared by %s and %s", ErrNameCollision, name, owner, normalized.ID)
			}
			nameOwners[name] = normalized.ID
		}
		fields = append(fields, normalized)
	}
	return fields, nil
}

func copyFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	cloned := make([]Field, len(fields))
	for i, field := range fields {
		cloned[i] = field
		cloned[i].Aliases = append([]string(nil), field.Aliases...)
	}
	return cloned
}
