package semantic

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTypeRef = errors.New("invalid semantic type reference")
	ErrUnknownType    = errors.New("unknown semantic type")
	ErrAmbiguousType  = errors.New("ambiguous semantic type")
	ErrTypeConflict   = errors.New("semantic type conflict")
)

// BuiltinTypeNamespace is the namespace reserved for the semantic kernel's
// built-in nominal types.
const BuiltinTypeNamespace Namespace = "gooo"

// BuiltinStringTypeID is the stable identity of the technology-independent
// semantic string type. It is intentionally not a Go type name.
const BuiltinStringTypeID ID = "urn:gooo:type:string"
const BuiltinStringTypeName = "string"

// TypeRef names a nominal semantic type. ID is the identity boundary. The
// optional Namespace and Name fields are lookup hints only and never
// participate in semantic identity.
type TypeRef struct {
	ID        ID
	Namespace Namespace
	Name      string
}

func NewTypeRef(id ID) (TypeRef, error) {
	ref := TypeRef{ID: id}
	return ref.Normalized()
}
func (r TypeRef) Normalized() (TypeRef, error) {
	if r.ID == "" && r.Name == "" {
		return TypeRef{}, fmt.Errorf("%w: ID or lookup name is required", ErrInvalidTypeRef)
	}
	if r.ID != "" {
		id, err := ParseIdentity(r.ID.String())
		if err != nil {
			return TypeRef{}, fmt.Errorf("%w: id: %v", ErrInvalidTypeRef, err)
		}
		r.ID = id
	}
	if r.Namespace != "" {
		namespace, err := ParseNamespace(r.Namespace.String())
		if err != nil {
			return TypeRef{}, fmt.Errorf("%w: namespace: %v", ErrInvalidTypeRef, err)
		}
		r.Namespace = namespace
	}
	if r.Name != "" {
		name, err := normalizeName(r.Name)
		if err != nil {
			return TypeRef{}, fmt.Errorf("%w: name: %v", ErrInvalidTypeRef, err)
		}
		r.Name = name
	}
	if r.ID == "" && r.Namespace != "" && r.Name == "" {
		return TypeRef{}, fmt.Errorf("%w: lookup namespace requires a name", ErrInvalidTypeRef)
	}
	return r, nil
}
func (r TypeRef) Validate() error {
	_, err := r.Normalized()
	return err
}
