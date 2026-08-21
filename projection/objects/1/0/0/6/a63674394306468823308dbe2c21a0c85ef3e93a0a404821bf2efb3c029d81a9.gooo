package bidir

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

var (
	ErrInvalidField         = errors.New("invalid bidir field")
	ErrUnrepresentableField = errors.New("bidir field is not representable")
)

// TypeRef is the nominal semantic type reference used by the BX carrier.
// Identity is supplied by the semantic registry; Name and Namespace are
// lookup spelling and never replace the stable ID.
type TypeRef = semantic.TypeRef
type FieldPresence = semantic.Presence
type FieldCardinality = semantic.Cardinality

const (
	FieldPresenceRequired = semantic.Required
	FieldPresenceOptional = semantic.Optional
	FieldCardinalityOne   = semantic.One
	FieldCardinalityMany  = semantic.Many
)

// Field is the latent, ordered BX carrier for one semantic field. Parent and
// ID are explicit identities; no display value participates in identity.
type Field struct {
	ID          ID
	Parent      ID
	Name        string
	Aliases     []string
	TypeRef     TypeRef
	TypeRefUse  TypeRefUse
	Origin      FieldOrigin
	Presence    FieldPresence
	Cardinality FieldCardinality
	Span        SourceSpan

	IDSpan          SourceSpan
	NameSpan        SourceSpan
	TypeRefSpan     SourceSpan
	PresenceSpan    SourceSpan
	CardinalitySpan SourceSpan
}

func (f Field) clone() Field {
	f.Aliases = append([]string(nil), f.Aliases...)
	return f
}
func cloneFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	cloned := make([]Field, len(fields))
	for index, field := range fields {
		cloned[index] = field.clone()
	}
	return cloned
}
