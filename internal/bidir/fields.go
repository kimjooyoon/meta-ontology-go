package bidir

import (
	"errors"
	"fmt"
	"strings"

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

func (f Field) semantic() (semantic.Field, error) {
	id, err := semantic.ParseIdentity(string(f.ID))
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: id: %v", ErrInvalidField, err)
	}
	parent, err := semantic.ParseIdentity(string(f.Parent))
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: parent: %v", ErrInvalidField, err)
	}
	typeRef := f.TypeRef
	if typeRef.ID == "" && typeRef.Name == "" && f.TypeRefUse.ResolvedID != "" {
		typeRef = semantic.TypeRef{ID: semantic.ID(f.TypeRefUse.ResolvedID)}
	}
	if typeRef.ID == "" && typeRef.Name == "" && f.TypeRefUse.Spelling != "" {
		var typeErr error
		typeRef, typeErr = typeRefFromUse(f.TypeRefUse)
		if typeErr != nil {
			return semantic.Field{}, fmt.Errorf("%w: type reference use: %v", ErrInvalidField, typeErr)
		}
	}
	field, err := (semantic.Field{
		ID:          id,
		Parent:      parent,
		Name:        f.Name,
		Aliases:     append([]string(nil), f.Aliases...),
		TypeRef:     typeRef,
		Presence:    f.Presence,
		Cardinality: f.Cardinality,
		Span:        toSemanticSpan(f.Span),
	}).Normalized()
	if err != nil {
		return semantic.Field{}, fmt.Errorf("%w: %v", ErrInvalidField, err)
	}
	return field, nil
}

func normalizeField(field Field, owner ID, kind Kind, registry semantic.TypeRegistry) (Field, error) {
	if kind != EntityKind {
		return Field{}, fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
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
		if err := item.span.Validate(); err != nil {
			return Field{}, fmt.Errorf("%w: %s span: %v", ErrInvalidField, item.name, err)
		}
	}
	if field.Origin != "" && !validFieldOrigin(field.Origin) {
		return Field{}, fmt.Errorf("%w: unknown field origin %q", ErrInvalidField, field.Origin)
	}
	use, err := normalizeTypeRefUse(field.TypeRefUse)
	if err != nil {
		return Field{}, fmt.Errorf("%w: type reference use: %v", ErrInvalidField, err)
	}
	normalized, err := field.semantic()
	if err != nil {
		return Field{}, err
	}
	if normalized.Parent != semantic.ID(owner) {
		return Field{}, fmt.Errorf("%w: field %s parent is %s, want %s", ErrInvalidField, normalized.ID, normalized.Parent, owner)
	}
	definition, err := resolveFieldType(normalized.TypeRef, use, registry)
	if err != nil {
		return Field{}, fmt.Errorf("%w: type ref: %v", ErrInvalidField, err)
	}
	result := field.clone()
	result.TypeRefUse = use
	result.TypeRefUse.ResolvedID = ID(definition.ID)
	return result, nil
}

func validateModelFields(nodes []Node, registry semantic.TypeRegistry) error {
	fieldOwners := make(map[ID]ID)
	nodeIDs := make(map[ID]Kind, len(nodes))
	for _, node := range nodes {
		nodeIDs[node.ID] = node.Kind
	}
	for _, node := range nodes {
		if len(node.Fields) == 0 {
			continue
		}
		if node.Kind != EntityKind {
			return fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
		}
		nameOwners := make(map[string]ID, len(node.Fields)*2)
		for index, field := range node.Fields {
			normalized, err := normalizeField(field, node.ID, node.Kind, registry)
			if err != nil {
				return fmt.Errorf("node %q field %d: %w", node.ID, index, err)
			}
			if owner, exists := fieldOwners[normalized.ID]; exists {
				return entityFieldsError(EntityFieldsIDCollisionDiagnostic, fmt.Sprintf("duplicate field ID %s owned by %s and %s", normalized.ID, owner, node.ID), field.Span, ErrInvalidField)
			}
			if kind, exists := nodeIDs[normalized.ID]; exists {
				return entityFieldsError(EntityFieldsIDCollisionDiagnostic, fmt.Sprintf("field ID %s collides with %s ID", normalized.ID, kind), field.Span, ErrInvalidField)
			}
			fieldOwners[normalized.ID] = node.ID
			for _, name := range append([]string{normalized.Name}, normalized.Aliases...) {
				if owner, exists := nameOwners[name]; exists && owner != normalized.ID {
					return fmt.Errorf("%w: field name %q is shared by %s and %s", ErrInvalidField, name, owner, normalized.ID)
				}
				nameOwners[name] = normalized.ID
			}
		}
	}
	return nil
}

func validateFieldOrderStability(before, after Model) error {
	beforeNodes, afterNodes := nodeMap(before.Nodes), nodeMap(after.Nodes)
	for id, source := range beforeNodes {
		updated, exists := afterNodes[id]
		if !exists || len(source.Fields) < 2 {
			continue
		}
		positions := make(map[ID]int, len(updated.Fields))
		for index, field := range updated.Fields {
			positions[field.ID] = index
		}
		last := -1
		for _, field := range source.Fields {
			position, exists := positions[field.ID]
			if !exists {
				continue
			}
			if position < last {
				return entityFieldsError(EntityFieldsIllegalReorderDiagnostic, fmt.Sprintf("field order for entity %s changed", id), field.Span, ErrInvalidField)
			}
			last = position
		}
	}
	return nil
}

func validateFieldParentStability(before, after Model) error {
	parents := make(map[ID]ID)
	for _, node := range before.Nodes {
		for _, field := range node.Fields {
			parents[field.ID] = field.Parent
		}
	}
	for _, node := range after.Nodes {
		for _, field := range node.Fields {
			if parent, exists := parents[field.ID]; exists && parent != field.Parent {
				return fmt.Errorf("%w: field %s cannot move from %s to %s", ErrInvalidField, field.ID, parent, field.Parent)
			}
		}
	}
	return nil
}

func fieldSemanticEqual(left, right Field) bool {
	leftSemantic, leftErr := left.semantic()
	rightSemantic, rightErr := right.semantic()
	if leftErr != nil || rightErr != nil {
		return false
	}
	registry := semantic.DefaultTypeRegistry()
	leftTypeID, leftErr := resolvedFieldTypeID(left, leftSemantic.TypeRef, registry)
	rightTypeID, rightErr := resolvedFieldTypeID(right, rightSemantic.TypeRef, registry)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftSemantic.ID == rightSemantic.ID &&
		leftSemantic.Parent == rightSemantic.Parent &&
		leftTypeID == rightTypeID &&
		leftSemantic.Presence == rightSemantic.Presence &&
		leftSemantic.Cardinality == rightSemantic.Cardinality
}

func writeFieldSemanticFingerprint(builder *strings.Builder, field Field) {
	normalized, err := field.semantic()
	if err != nil {
		writeFingerprintPart(builder, "invalid:"+string(field.ID))
		return
	}
	typeID, err := resolvedFieldTypeID(field, normalized.TypeRef, semantic.DefaultTypeRegistry())
	if err != nil {
		writeFingerprintPart(builder, "invalid-type:"+string(field.ID))
		return
	}
	writeFingerprintPart(builder, string(normalized.ID))
	writeFingerprintPart(builder, string(normalized.Parent))
	writeFingerprintPart(builder, string(typeID))
	writeFingerprintPart(builder, string(normalized.Presence))
	writeFingerprintPart(builder, string(normalized.Cardinality))
}

func resolvedTypeID(ref semantic.TypeRef, registry semantic.TypeRegistry) (semantic.ID, error) {
	normalized, err := ref.Normalized()
	if err != nil {
		return "", err
	}
	if normalized.ID != "" {
		return normalized.ID, nil
	}
	definition, err := registry.Resolve(normalized)
	if err != nil {
		return "", err
	}
	return definition.ID, nil
}
