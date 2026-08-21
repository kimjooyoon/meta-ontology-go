package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func resolveFieldType(semanticRef semantic.TypeRef, use TypeRefUse, registry semantic.TypeRegistry) (semantic.TypeDef, error) {
	current, err := registry.Resolve(semanticRef)
	if semanticRef.ID != "" {
		return resolveExplicitFieldType(current, err, use, registry)
	}
	if err != nil && use.ResolvedID == "" {
		return semantic.TypeDef{}, err
	}
	if presentationErr := validateFieldLookupPresentation(semanticRef, use, err); presentationErr != nil {
		return semantic.TypeDef{}, presentationErr
	}
	if use.Form == TypeRefFormLookup && use.ResolvedID == "" {
		return resolveLookupFieldType(current, err, use, registry)
	}
	if use.ResolvedID != "" {
		return resolveFieldTypeByID(current, err, use, registry)
	}
	return current, err
}
func resolveExplicitFieldType(current semantic.TypeDef, resolveErr error, use TypeRefUse, registry semantic.TypeRegistry) (semantic.TypeDef, error) {
	if resolveErr != nil {
		return semantic.TypeDef{}, resolveErr
	}
	if use.Form == TypeRefFormLookup && use.ResolvedID == "" {
		presentationRef, err := parseLookupTypeRef(use.Spelling)
		if err != nil {
			return semantic.TypeDef{}, err
		}
		presented, err := registry.Resolve(presentationRef)
		if err != nil {
			return semantic.TypeDef{}, err
		}
		if current.ID != presented.ID {
			return semantic.TypeDef{}, fmt.Errorf("field TypeRef and source presentation resolve to different IDs (%s and %s)", current.ID, presented.ID)
		}
	}
	if use.ResolvedID == "" {
		return current, nil
	}
	resolvedID, err := semantic.ParseIdentity(string(use.ResolvedID))
	if err != nil {
		return semantic.TypeDef{}, err
	}
	if current.ID != resolvedID {
		return semantic.TypeDef{}, fmt.Errorf("explicit field TypeRef ID %s disagrees with source presentation ID %s", current.ID, resolvedID)
	}
	return current, nil
}
