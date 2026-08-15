package semantic

import (
	"errors"
	"fmt"
	"sort"
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

func (r *TypeRegistry) Register(def TypeDef) error {
	normalized, err := def.Normalized()
	if err != nil {
		return err
	}
	r.ensure()
	if existing, ok := r.types[normalized.ID]; ok {
		if existing != normalized {
			return fmt.Errorf("%w: %s is already registered", ErrTypeConflict, normalized.ID)
		}
		return nil
	}
	r.types[normalized.ID] = normalized
	ref := NameRef{Namespace: normalized.Namespace, Name: normalized.Name}
	for _, id := range r.names[ref] {
		if id == normalized.ID {
			return nil
		}
	}
	r.names[ref] = append(r.names[ref], normalized.ID)
	sort.Slice(r.names[ref], func(i, j int) bool { return r.names[ref][i] < r.names[ref][j] })
	return nil
}

func (r TypeRegistry) Resolve(ref TypeRef) (TypeDef, error) {
	normalized, err := ref.Normalized()
	if err != nil {
		return TypeDef{}, err
	}
	if normalized.ID != "" {
		def, ok := r.types[normalized.ID]
		if !ok {
			return TypeDef{}, fmt.Errorf("%w: %s", ErrUnknownType, normalized.ID)
		}
		if normalized.Name != "" && normalized.Name != def.Name {
			return TypeDef{}, fmt.Errorf("%w: lookup name %q does not match %s", ErrInvalidTypeRef, normalized.Name, def.ID)
		}
		if normalized.Namespace != "" && normalized.Namespace != def.Namespace {
			return TypeDef{}, fmt.Errorf("%w: lookup namespace %q does not match %s", ErrInvalidTypeRef, normalized.Namespace, def.ID)
		}
		return def, nil
	}

	ids := r.lookupName(normalized.Namespace, normalized.Name)
	switch len(ids) {
	case 0:
		return TypeDef{}, fmt.Errorf("%w: %s", ErrUnknownType, normalized.Name)
	case 1:
		return r.types[ids[0]], nil
	default:
		return TypeDef{}, fmt.Errorf("%w: %s", ErrAmbiguousType, normalized.Name)
	}
}

func (r TypeRegistry) ResolveName(namespace Namespace, name string) (TypeDef, error) {
	return r.Resolve(TypeRef{Namespace: namespace, Name: name})
}

func (r TypeRegistry) lookupName(namespace Namespace, name string) []ID {
	if namespace != "" {
		ref, err := NewNameRef(namespace, name)
		if err != nil {
			return nil
		}
		return append([]ID(nil), r.names[ref]...)
	}
	ids := make([]ID, 0)
	for ref, candidates := range r.names {
		if ref.Name != name {
			continue
		}
		ids = append(ids, candidates...)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	unique := ids[:0]
	for _, id := range ids {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return unique
}
