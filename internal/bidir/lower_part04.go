package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func parseSyntaxTypeRef(spelling string) (semantic.TypeRef, TypeRefUse, error) {
	raw := strings.TrimSpace(spelling)
	if raw == "" {
		return semantic.TypeRef{}, TypeRefUse{}, fmt.Errorf("%w: type reference spelling is empty", ErrInvalidField)
	}
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "urn:") {
		id, err := semantic.ParseIdentity(raw)
		if err != nil {
			return semantic.TypeRef{}, TypeRefUse{}, fmt.Errorf("%w: type reference identity: %v", ErrInvalidField, err)
		}
		return semantic.TypeRef{ID: id}, TypeRefUse{Form: TypeRefFormStableID, Spelling: id.String(), ResolvedID: ID(id)}, nil
	}
	ref, err := parseLookupTypeRef(raw)
	if err != nil {
		return semantic.TypeRef{}, TypeRefUse{}, err
	}
	return ref, TypeRefUse{Form: TypeRefFormLookup, Spelling: lookupTypeRefSpelling(ref)}, nil
}
func parseLookupTypeRef(raw string) (semantic.TypeRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference spelling is empty", ErrInvalidField)
	}
	parts := strings.Split(raw, ":")
	ref := semantic.TypeRef{}
	switch len(parts) {
	case 1:
		ref.Name = strings.TrimSpace(parts[0])
	case 2:
		namespace, err := semantic.ParseNamespace(strings.TrimSpace(parts[0]))
		if err != nil {
			return semantic.TypeRef{}, fmt.Errorf("%w: type reference namespace: %v", ErrInvalidField, err)
		}
		ref.Namespace = namespace
		ref.Name = strings.TrimSpace(parts[1])
	default:
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference spelling %q is not representable", ErrInvalidField, raw)
	}
	if err := ref.Validate(); err != nil {
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference: %v", ErrInvalidField, err)
	}
	return ref, nil
}
