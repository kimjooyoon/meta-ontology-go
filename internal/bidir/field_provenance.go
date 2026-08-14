package bidir

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// FieldOrigin identifies the provenance class of a latent field carrier.
// Only Source values are authoritative for public source readback.
type FieldOrigin string

const (
	FieldOriginSource      FieldOrigin = "Source"
	FieldOriginGenerated   FieldOrigin = "Generated"
	FieldOriginSynthesized FieldOrigin = "Synthesized"
	FieldOriginDeferred    FieldOrigin = "Deferred"
	FieldOriginUnsupported FieldOrigin = "Unsupported"
)

// TypeRefForm records the source/BX spelling form of a nominal type use.
// Form and Spelling are presentation provenance; ResolvedID is the semantic
// identity resolved from that presentation.
type TypeRefForm string

const (
	TypeRefFormLookup   TypeRefForm = "Lookup"
	TypeRefFormStableID TypeRefForm = "StableID"
)

// TypeRefUse preserves the original source spelling independently from the
// semantic TypeRef. It must never be reconstructed from a registry name.
type TypeRefUse struct {
	Form       TypeRefForm
	Spelling   string
	ResolvedID ID
	Span       SourceSpan
}

func validFieldOrigin(origin FieldOrigin) bool {
	switch origin {
	case FieldOriginSource, FieldOriginGenerated, FieldOriginSynthesized, FieldOriginDeferred, FieldOriginUnsupported:
		return true
	default:
		return false
	}
}

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

func resolveFieldType(semanticRef semantic.TypeRef, use TypeRefUse, registry semantic.TypeRegistry) (semantic.TypeDef, error) {
	current, err := registry.Resolve(semanticRef)
	if err != nil && use.ResolvedID == "" {
		return semantic.TypeDef{}, err
	}
	if use.Form == TypeRefFormLookup && use.ResolvedID != "" && semanticRef.ID == "" {
		presentationRef, presentationErr := parseLookupTypeRef(use.Spelling)
		if presentationErr != nil {
			return semantic.TypeDef{}, presentationErr
		}
		currentRef, currentErr := semanticRef.Normalized()
		if currentErr != nil {
			return semantic.TypeDef{}, currentErr
		}
		if currentRef.Name != presentationRef.Name || (currentRef.Namespace != "" && presentationRef.Namespace != "" && currentRef.Namespace != presentationRef.Namespace) {
			if err != nil {
				return semantic.TypeDef{}, err
			}
			return semantic.TypeDef{}, fmt.Errorf("field TypeRef lookup does not match its original source presentation")
		}
	}
	if use.Form == TypeRefFormLookup && use.ResolvedID == "" {
		presentationRef, presentationErr := typeRefFromUse(use)
		if presentationErr != nil {
			return semantic.TypeDef{}, presentationErr
		}
		presented, presentationErr := registry.Resolve(presentationRef)
		if presentationErr != nil {
			return semantic.TypeDef{}, presentationErr
		}
		if err != nil {
			return semantic.TypeDef{}, err
		}
		if current.ID != presented.ID {
			return semantic.TypeDef{}, fmt.Errorf("field TypeRef and source presentation resolve to different IDs (%s and %s)", current.ID, presented.ID)
		}
		current = presented
	}
	if use.ResolvedID != "" {
		resolved, resolveErr := registry.Resolve(semantic.TypeRef{ID: semantic.ID(use.ResolvedID)})
		if resolveErr != nil {
			return semantic.TypeDef{}, resolveErr
		}
		if err == nil && current.ID != resolved.ID {
			return semantic.TypeDef{}, fmt.Errorf("field TypeRef and source presentation resolve to different IDs (%s and %s)", current.ID, resolved.ID)
		}
		return resolved, nil
	}
	return current, err
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

func resolvedFieldTypeID(field Field, ref semantic.TypeRef, registry semantic.TypeRegistry) (semantic.ID, error) {
	if field.TypeRefUse.ResolvedID != "" {
		id, err := semantic.ParseIdentity(string(field.TypeRefUse.ResolvedID))
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return resolvedTypeID(ref, registry)
}

func validateSourceField(field Field, owner ID, registry semantic.TypeRegistry) error {
	if field.Origin != FieldOriginSource {
		return fmt.Errorf("%w: field %q has non-source origin %q", ErrUnrepresentableField, field.ID, field.Origin)
	}
	spans := []struct {
		name string
		span SourceSpan
	}{
		{name: "field", span: field.Span},
		{name: "field ID", span: field.IDSpan},
		{name: "field name", span: field.NameSpan},
		{name: "field type", span: field.TypeRefSpan},
		{name: "field presence", span: field.PresenceSpan},
		{name: "field cardinality", span: field.CardinalitySpan},
	}
	for _, item := range spans {
		if !item.span.Valid() {
			return fmt.Errorf("%w: %s span is not exact source provenance", ErrUnrepresentableField, item.name)
		}
	}
	normalizedUse, err := normalizeTypeRefUse(field.TypeRefUse)
	if err != nil {
		return fmt.Errorf("%w: type reference presentation: %v", ErrUnrepresentableField, err)
	}
	if normalizedUse.Form == "" || normalizedUse.Spelling == "" {
		return fmt.Errorf("%w: field %q has no original type reference presentation", ErrUnrepresentableField, field.ID)
	}
	if normalizedUse.Spelling != field.TypeRefUse.Spelling {
		return fmt.Errorf("%w: field %q type reference spelling is not normalized", ErrUnrepresentableField, field.ID)
	}
	if normalizedUse.Span != field.TypeRefSpan {
		return fmt.Errorf("%w: field %q type reference span does not match its source subspan", ErrUnrepresentableField, field.ID)
	}
	normalized, err := normalizeField(field, owner, EntityKind, registry)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnrepresentableField, err)
	}
	if field.TypeRefUse.ResolvedID != "" && field.TypeRefUse.ResolvedID != normalized.TypeRefUse.ResolvedID {
		return fmt.Errorf("%w: field %q source TypeRef resolved ID changed", ErrUnrepresentableField, field.ID)
	}
	return nil
}
