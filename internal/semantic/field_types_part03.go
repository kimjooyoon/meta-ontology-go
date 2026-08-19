package semantic

import (
	"fmt"
	"sort"
)

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
