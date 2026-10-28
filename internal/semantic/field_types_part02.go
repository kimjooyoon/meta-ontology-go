package semantic

import (
	"fmt"
	"maps"
	"slices"
)

// TypeDef describes a registered nominal semantic type. Name is a lookup
// convenience; callers must retain and exchange ID as the stable reference.
type TypeDef struct {
	ID        ID
	Namespace Namespace
	Name      string
}

func (d TypeDef) Normalized() (TypeDef, error) {
	id, err := ParseIdentity(d.ID.String())
	if err != nil {
		return TypeDef{}, fmt.Errorf("%w: id: %v", ErrInvalidTypeRef, err)
	}
	namespace, err := ParseNamespace(d.Namespace.String())
	if err != nil {
		return TypeDef{}, fmt.Errorf("%w: namespace: %v", ErrInvalidTypeRef, err)
	}
	name, err := normalizeName(d.Name)
	if err != nil {
		return TypeDef{}, fmt.Errorf("%w: name: %v", ErrInvalidTypeRef, err)
	}
	d.ID = id
	d.Namespace = namespace
	d.Name = name
	return d, nil
}
func (d TypeDef) Validate() error {
	_, err := d.Normalized()
	return err
}
func BuiltinStringType() TypeDef {
	return TypeDef{ID: BuiltinStringTypeID, Namespace: BuiltinTypeNamespace, Name: BuiltinStringTypeName}
}
func BuiltinBooleanType() TypeDef {
	return TypeDef{ID: BuiltinBooleanTypeID, Namespace: BuiltinTypeNamespace, Name: BuiltinBooleanTypeName}
}
func BuiltinIntegerType() TypeDef {
	return TypeDef{ID: BuiltinIntegerTypeID, Namespace: BuiltinTypeNamespace, Name: BuiltinIntegerTypeName}
}

// BuiltInStringType is an alternate spelling for callers using the usual
// English capitalization of "built-in".
func BuiltInStringType() TypeDef {
	return BuiltinStringType()
}

// TypeRegistry stores nominal type definitions and lookup indexes. The zero
// value is usable for registration; NewTypeRegistry includes the built-in
// semantic string type.
type TypeRegistry struct {
	types map[ID]TypeDef
	names map[NameRef][]ID
}

func NewTypeRegistry() TypeRegistry {
	registry := TypeRegistry{}
	if err := registry.Register(BuiltinStringType()); err != nil {
		panic(err)
	}
	if err := registry.Register(BuiltinBooleanType()); err != nil {
		panic(err)
	}
	if err := registry.Register(BuiltinIntegerType()); err != nil {
		panic(err)
	}
	return registry
}
func DefaultTypeRegistry() TypeRegistry {
	return NewTypeRegistry()
}
func (r *TypeRegistry) ensure() {
	if r.types == nil {
		r.types = make(map[ID]TypeDef)
	}
	if r.names == nil {
		r.names = make(map[NameRef][]ID)
	}
}

// Clone detaches the registry so a source compilation cannot register types in
// a caller's registry, including when a later declaration conflicts.
func (r TypeRegistry) Clone() TypeRegistry {
	result := TypeRegistry{types: maps.Clone(r.types), names: make(map[NameRef][]ID, len(r.names))}
	for name, ids := range r.names {
		result.names[name] = slices.Clone(ids)
	}
	return result
}
