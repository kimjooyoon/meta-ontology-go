package semantic

import (
	"errors"
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
