package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateFieldLookupPresentation(ref semantic.TypeRef, use TypeRefUse, resolveErr error) error {
	if use.Form != TypeRefFormLookup || use.ResolvedID == "" {
		return nil
	}
	presentationRef, err := parseLookupTypeRef(use.Spelling)
	if err != nil {
		return err
	}
	currentRef, err := ref.Normalized()
	if err != nil {
		return err
	}
	if currentRef.Name == presentationRef.Name && (currentRef.Namespace == "" || presentationRef.Namespace == "" || currentRef.Namespace == presentationRef.Namespace) {
		return nil
	}
	if resolveErr != nil {
		return resolveErr
	}
	return fmt.Errorf("field TypeRef lookup does not match its original source presentation")
}
func resolveLookupFieldType(current semantic.TypeDef, resolveErr error, use TypeRefUse, registry semantic.TypeRegistry) (semantic.TypeDef, error) {
	presentationRef, err := typeRefFromUse(use)
	if err != nil {
		return semantic.TypeDef{}, err
	}
	presented, err := registry.Resolve(presentationRef)
	if err != nil {
		return semantic.TypeDef{}, err
	}
	if resolveErr != nil {
		return semantic.TypeDef{}, resolveErr
	}
	if current.ID != presented.ID {
		return semantic.TypeDef{}, fmt.Errorf("field TypeRef and source presentation resolve to different IDs (%s and %s)", current.ID, presented.ID)
	}
	return presented, nil
}
func resolveFieldTypeByID(current semantic.TypeDef, resolveErr error, use TypeRefUse, registry semantic.TypeRegistry) (semantic.TypeDef, error) {
	resolved, err := registry.Resolve(semantic.TypeRef{ID: semantic.ID(use.ResolvedID)})
	if err != nil {
		return semantic.TypeDef{}, err
	}
	if resolveErr == nil && current.ID != resolved.ID {
		return semantic.TypeDef{}, fmt.Errorf("field TypeRef and source presentation resolve to different IDs (%s and %s)", current.ID, resolved.ID)
	}
	return resolved, nil
}
func normalizeModelFields(model *Model, registry semantic.TypeRegistry) error {
	for nodeIndex := range model.Nodes {
		for fieldIndex := range model.Nodes[nodeIndex].Fields {
			normalized, err := normalizeField(model.Nodes[nodeIndex].Fields[fieldIndex], model.Nodes[nodeIndex].ID, model.Nodes[nodeIndex].Kind, registry)
			if err != nil {
				return fmt.Errorf("node %q field %d: %w", model.Nodes[nodeIndex].ID, fieldIndex, err)
			}
			model.Nodes[nodeIndex].Fields[fieldIndex].TypeRefUse = normalized.TypeRefUse
		}
	}
	return nil
}
