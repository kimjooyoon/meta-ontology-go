package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func normalizeTypeRefUse(use TypeRefUse) (TypeRefUse, error) {
	if use.Form == "" && use.Spelling == "" && use.ResolvedID == "" && !use.Span.Valid() {
		return TypeRefUse{}, nil
	}
	raw := strings.TrimSpace(use.Spelling)
	if raw == "" {
		return TypeRefUse{}, fmt.Errorf("type reference presentation is missing")
	}
	switch use.Form {
	case TypeRefFormLookup:
		ref, err := parseLookupTypeRef(raw)
		if err != nil {
			return TypeRefUse{}, err
		}
		use.Spelling = lookupTypeRefSpelling(ref)
	case TypeRefFormStableID:
		id, err := semantic.ParseIdentity(raw)
		if err != nil {
			return TypeRefUse{}, fmt.Errorf("stable type reference identity: %v", err)
		}
		use.Spelling = id.String()
		if use.ResolvedID != "" && use.ResolvedID != ID(id) {
			return TypeRefUse{}, fmt.Errorf("resolved type reference ID %q disagrees with spelling %q", use.ResolvedID, id)
		}
		use.ResolvedID = ID(id)
	default:
		return TypeRefUse{}, fmt.Errorf("unknown type reference form %q", use.Form)
	}
	if use.ResolvedID != "" {
		id, err := semantic.ParseIdentity(string(use.ResolvedID))
		if err != nil {
			return TypeRefUse{}, fmt.Errorf("resolved type reference ID: %v", err)
		}
		use.ResolvedID = ID(id)
	}
	return use, nil
}
func typeRefFromUse(use TypeRefUse) (semantic.TypeRef, error) {
	normalized, err := normalizeTypeRefUse(use)
	if err != nil {
		return semantic.TypeRef{}, err
	}
	if normalized.ResolvedID != "" {
		return semantic.TypeRef{ID: semantic.ID(normalized.ResolvedID)}, nil
	}
	return parseLookupTypeRef(normalized.Spelling)
}
func lookupTypeRefSpelling(ref semantic.TypeRef) string {
	if ref.Namespace != "" {
		return ref.Namespace.String() + ":" + ref.Name
	}
	return ref.Name
}
