package generator

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	entityFieldsDeferredDiagnostic         = "parse.entity-fields-deferred"
	entityFieldsConfigurationDiagnostic    = "GOOO-EF-V1-CONFIGURATION"
	entityFieldsUnboundProfileDiagnostic   = "GOOO-EF-V1-UNBOUND-PROFILE"
	entityFieldsProfileMismatchDiagnostic  = "GOOO-EF-V1-PROFILE-MISMATCH"
	entityFieldsProfileDigestDiagnostic    = "GOOO-EF-V1-PROFILE-DIGEST-MISMATCH"
	entityFieldsUnknownStateDiagnostic     = "GOOO-EF-V1-UNKNOWN-STATE"
	entityFieldsUnknownTypeDiagnostic      = "GOOO-EF-V1-UNKNOWN-TYPE"
	entityFieldsUnsupportedTypeDiagnostic  = "GOOO-EF-V1-UNSUPPORTED-TYPE"
	entityFieldsUnsupportedShapeDiagnostic = "GOOO-EF-V1-UNSUPPORTED-SHAPE"
	entityFieldsIDCollisionDiagnostic      = "GOOO-EF-V1-ID-COLLISION"
	entityFieldsWrongParentDiagnostic      = "GOOO-EF-V1-WRONG-PARENT"
	entityFieldsIncompleteDiagnostic       = "GOOO-EF-V1-INCOMPLETE-FIELD"
	entityFieldsIllegalReorderDiagnostic   = "GOOO-EF-V1-ILLEGAL-REORDER"
	entityFieldsGoNameCollisionDiagnostic  = "GOOO-EF-V1-GO-NAME-COLLISION"
	entityFieldsUnrepresentableDiagnostic  = "GOOO-EF-V1-UNREPRESENTABLE"
)

const entityFieldsStringTypeID = "urn:gooo:type:string"

type entityFieldsSupport = syntax.EntityFieldsSupport

const entityFieldsSupported = syntax.EntityFieldsSupported

type entityFieldsProjectionError struct {
	code  string
	field Field
	text  string
}

func (e *entityFieldsProjectionError) Error() string {
	if e == nil {
		return entityFieldsConfigurationDiagnostic
	}
	message := "generator: " + e.code
	if e.field.ID != "" {
		message += fmt.Sprintf(": field %q", e.field.ID)
	}
	if e.field.Source.URI != "" {
		message += fmt.Sprintf(" at %s:%d:%d", e.field.Source.URI, e.field.Source.Start.Line, e.field.Source.Start.Column)
	}
	if e.text != "" {
		message += ": " + e.text
	}
	return message
}

func entityFieldsError(code string, field Field, text string) error {
	return &entityFieldsProjectionError{code: code, field: field, text: text}
}

func checkedEntityFieldsSupport() syntax.EntityFieldsSupport {
	return syntax.CurrentEntityFieldsSupport()
}

func validateEntityFieldsSupport(support syntax.EntityFieldsSupport) error {
	switch support.State {
	case syntax.EntityFieldsDeferred, syntax.EntityFieldsSupported:
	default:
		return entityFieldsError(entityFieldsUnknownStateDiagnostic, Field{}, fmt.Sprintf("unknown support state %q", support.State))
	}
	if support.Profile.ID == "" && support.Profile.Version == 0 && support.Profile.Digest == "" {
		return entityFieldsError(entityFieldsUnboundProfileDiagnostic, Field{}, "profile is unbound")
	}
	if support.Profile.ID != syntax.EntityFieldsProfileID || support.Profile.Version != syntax.EntityFieldsProfileVersion {
		return entityFieldsError(entityFieldsProfileMismatchDiagnostic, Field{}, "profile identity or version does not match")
	}
	if support.Profile.Digest != syntax.EntityFieldsProfileDigest {
		return entityFieldsError(entityFieldsProfileDigestDiagnostic, Field{}, "profile digest does not match")
	}
	return nil
}

func validateEntityFieldsInput(ir SemanticIR, support syntax.EntityFieldsSupport) error {
	if err := validateEntityFieldsSupport(support); err != nil {
		return err
	}
	if !semanticIRHasFields(ir) {
		return nil
	}
	first := firstSemanticIRField(ir)
	if support.State == syntax.EntityFieldsDeferred {
		return entityFieldsError(entityFieldsDeferredDiagnostic, first, "entity fields are deferred and unsupported by the public generator")
	}
	return validateSupportedEntityFields(ir)
}

func semanticIRHasFields(ir SemanticIR) bool {
	for _, entity := range ir.Entities {
		if len(entity.Fields) > 0 {
			return true
		}
	}
	return false
}

func firstSemanticIRField(ir SemanticIR) Field {
	for _, entity := range ir.Entities {
		if len(entity.Fields) > 0 {
			return entity.Fields[0]
		}
	}
	return Field{}
}

func validateSupportedEntityFields(ir SemanticIR) error {
	used := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		if previous, exists := used[entity.ID]; exists {
			return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", entity.ID, previous))
		}
		used[entity.ID] = "entity"
	}
	for _, activity := range ir.Activities {
		if previous, exists := used[activity.ID]; exists {
			return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", activity.ID, previous))
		}
		used[activity.ID] = "activity"
		for _, slot := range activity.Slots {
			if previous, exists := used[slot.ID]; exists {
				return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", slot.ID, previous))
			}
			used[slot.ID] = "slot"
		}
	}

	for _, entity := range ir.Entities {
		if len(entity.Fields) == 0 {
			continue
		}
		seenNames := make(map[string]string, len(entity.Fields))
		var sourceURI string
		var previousStart int
		var hasPrevious bool
		for index, field := range entity.Fields {
			if err := validateSupportedField(entity, index, field, used, seenNames, sourceURI, previousStart, hasPrevious); err != nil {
				return err
			}
			if sourceURI == "" {
				sourceURI = field.Source.URI
			}
			previousStart = field.Source.Start.Offset
			hasPrevious = true
		}
	}
	return nil
}

func validateSupportedField(entity Entity, index int, field Field, used map[string]string, seenNames map[string]string, sourceURI string, previousStart int, hasPrevious bool) error {
	context := fmt.Sprintf("entity %q field %d", entity.ID, index)
	if field.ID == "" || field.Name == "" || field.Parent == "" || field.TypeRefID == "" || field.Presence == "" || field.Cardinality == "" {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, context+" is missing a required identity, name, type, shape, or parent")
	}
	canonicalID, err := canonicalEntityFieldIdentity(field.ID)
	if err != nil || canonicalID != field.ID {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, "field ID must be a canonical absolute identity")
	}
	if previous, exists := used[field.ID]; exists {
		return entityFieldsError(entityFieldsIDCollisionDiagnostic, field, fmt.Sprintf("identity is already used by %s", previous))
	}
	used[field.ID] = "field"
	if field.Parent != entity.ID {
		return entityFieldsError(entityFieldsWrongParentDiagnostic, field, fmt.Sprintf("parent %q does not match entity %q", field.Parent, entity.ID))
	}
	if len(field.Aliases) != 0 {
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, field, "aliases are not part of the bound projection profile")
	}
	if field.Origin != "" && field.Origin != "Source" {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, fmt.Sprintf("field origin %q is not source-authoritative", field.Origin))
	}
	if field.GoName != "" && field.GoName != field.Name {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "Go name must be exactly Field.Name")
	}
	if field.GoType != "" && field.GoType != "string" {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "Go type is derived from the profile and cannot be overridden")
	}
	if !isGoIdentifier(field.Name) {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, "Field.Name is not a valid Go identifier")
	}
	if previous, exists := seenNames[field.Name]; exists {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, fmt.Sprintf("Go name is already used by field %q", previous))
	}
	seenNames[field.Name] = field.ID
	if field.TypeRefID != entityFieldsStringTypeID {
		return entityFieldsError(entityFieldsUnsupportedTypeDiagnostic, field, fmt.Sprintf("resolved type %q is not in the bound profile", field.TypeRefID))
	}
	if field.Presence != "required" || field.Cardinality != "one" {
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, field, "only required × one is in the bound profile")
	}
	if err := validateFieldSource(entity, field, sourceURI, previousStart, hasPrevious); err != nil {
		return err
	}
	return nil
}

func validateFieldSource(entity Entity, field Field, sourceURI string, previousStart int, hasPrevious bool) error {
	span := field.Source
	if strings.TrimSpace(span.URI) == "" || span.Start.Line <= 0 || span.Start.Column <= 0 || span.End.Line <= 0 || span.End.Column <= 0 || span.Start.Offset < 0 || span.End.Offset <= span.Start.Offset {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, "source span must be a non-zero, half-open source range")
	}
	if sourceURI != "" && span.URI != sourceURI {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "field origins cross source snapshots")
	}
	if hasPrevious && span.Start.Offset <= previousStart {
		return entityFieldsError(entityFieldsIllegalReorderDiagnostic, field, fmt.Sprintf("%s field order is not source ordered", entity.ID))
	}
	if entity.Source.URI != "" && entity.Source.URI != span.URI {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "field origin does not match its entity source snapshot")
	}
	for _, subspan := range []struct {
		name string
		span SourceSpan
	}{
		{name: "ID", span: field.IDSpan},
		{name: "name", span: fieldNameSource(field)},
		{name: "type", span: field.TypeRefSpan},
		{name: "presence", span: field.PresenceSpan},
		{name: "cardinality", span: field.CardinalitySpan},
	} {
		if !sourceSpanIsZero(subspan.span) {
			if err := validateFieldSubspan(field, subspan.span, subspan.name); err != nil {
				return err
			}
		}
	}
	if !sourceSpanIsZero(field.NameSpan) && !sourceSpanIsZero(field.NameSource) && field.NameSpan != field.NameSource {
		if err := validateFieldSubspan(field, field.NameSpan, "name"); err != nil {
			return err
		}
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "name source spans disagree")
	}
	return nil
}

func fieldNameSource(field Field) SourceSpan {
	if !sourceSpanIsZero(field.NameSpan) {
		return field.NameSpan
	}
	return field.NameSource
}

func sourceSpanIsZero(span SourceSpan) bool {
	return span == (SourceSpan{})
}

func validateFieldSubspan(field Field, subspan SourceSpan, label string) error {
	if strings.TrimSpace(subspan.URI) == "" || subspan.URI != field.Source.URI || subspan.Start.Line <= 0 || subspan.Start.Column <= 0 || subspan.End.Line <= 0 || subspan.End.Column <= 0 || subspan.Start.Offset < field.Source.Start.Offset || subspan.End.Offset > field.Source.End.Offset || subspan.End.Offset <= subspan.Start.Offset {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, label+" source span is not an exact subspan of the field origin")
	}
	return nil
}

func canonicalEntityFieldIdentity(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("empty or spaced identity")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || strings.Contains(value, "://") && parsed.Host == "" {
		return "", fmt.Errorf("identity is not absolute")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Host != "" {
		parsed.Host = strings.ToLower(parsed.Host)
	}
	return parsed.String(), nil
}

func prepareEntityFields(ir SemanticIR) SemanticIR {
	prepared := copyIR(ir)
	for entityIndex := range prepared.Entities {
		for fieldIndex := range prepared.Entities[entityIndex].Fields {
			field := &prepared.Entities[entityIndex].Fields[fieldIndex]
			field.GoName = field.Name
			field.GoType = "string"
		}
	}
	return prepared
}

func entityFieldsMetadata(support syntax.EntityFieldsSupport) *EntityFieldsMetadata {
	return &EntityFieldsMetadata{
		State: string(support.State),
		Profile: EntityFieldsProfileMetadata{
			ID: support.Profile.ID, Version: support.Profile.Version, Digest: support.Profile.Digest,
		},
	}
}

func entityFieldsProfileMapping() syntax.EntityFieldsProfile {
	return syntax.CurrentEntityFieldsSupport().Profile
}
