package bidir

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	EntityFieldsDeferredDiagnostic         = "parse.entity-fields-deferred"
	EntityFieldsConfigurationDiagnostic    = "parse.entity-fields-configuration"
	EntityFieldsUnboundProfileDiagnostic   = "GOOO-EF-V1-UNBOUND-PROFILE"
	EntityFieldsProfileDigestDiagnostic    = "GOOO-EF-V1-PROFILE-DIGEST-MISMATCH"
	EntityFieldsProfileMismatchDiagnostic  = "GOOO-EF-V1-PROFILE-MISMATCH"
	EntityFieldsUnknownStateDiagnostic     = "GOOO-EF-V1-UNKNOWN-STATE"
	EntityFieldsUnknownTypeDiagnostic      = "GOOO-EF-V1-UNKNOWN-TYPE"
	EntityFieldsAmbiguousTypeDiagnostic    = "GOOO-EF-V1-AMBIGUOUS-TYPE"
	EntityFieldsUnsupportedTypeDiagnostic  = "GOOO-EF-V1-UNSUPPORTED-TYPE"
	EntityFieldsUnsupportedShapeDiagnostic = "GOOO-EF-V1-UNSUPPORTED-SHAPE"
	EntityFieldsIDCollisionDiagnostic      = "GOOO-EF-V1-ID-COLLISION"
	EntityFieldsWrongParentDiagnostic      = "GOOO-EF-V1-WRONG-PARENT"
	EntityFieldsIncompleteDiagnostic       = "GOOO-EF-V1-INCOMPLETE-FIELD"
	EntityFieldsIllegalReorderDiagnostic   = "GOOO-EF-V1-ILLEGAL-REORDER"
)

type EntityFieldsState = syntax.EntityFieldsState
type EntityFieldsProfile = syntax.EntityFieldsProfile
type EntityFieldsSupport = syntax.EntityFieldsSupport

const (
	EntityFieldsDeferred  = syntax.EntityFieldsDeferred
	EntityFieldsSupported = syntax.EntityFieldsSupported
)

var (
	ErrEntityFieldsDeferred        = errors.New(EntityFieldsDeferredDiagnostic)
	ErrEntityFieldsConfiguration   = errors.New(EntityFieldsConfigurationDiagnostic)
	ErrEntityFieldsUnboundProfile  = errors.New(EntityFieldsUnboundProfileDiagnostic)
	ErrEntityFieldsProfileDigest   = errors.New(EntityFieldsProfileDigestDiagnostic)
	ErrEntityFieldsProfileMismatch = errors.New(EntityFieldsProfileMismatchDiagnostic)
	ErrEntityFieldsUnknownState    = errors.New(EntityFieldsUnknownStateDiagnostic)
)

// EntityFieldsError is the stable bidir error for profile and activation
// failures. Span is retained so callers can attach the source-backed failure
// to the originating field without making source spans semantic identity.
type EntityFieldsError struct {
	Code    string
	Message string
	Span    SourceSpan
	Cause   error
}

func (e *EntityFieldsError) Error() string {
	if e == nil {
		return EntityFieldsConfigurationDiagnostic
	}
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func (e *EntityFieldsError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func CurrentEntityFieldsSupport() EntityFieldsSupport {
	return syntax.CurrentEntityFieldsSupport()
}

// supportedEntityFieldsForTest returns the exact profile-bound SUPPORTED
// branch for explicit tests. No process-global support state is changed.
func supportedEntityFieldsForTest() EntityFieldsSupport {
	support := CurrentEntityFieldsSupport()
	support.State = EntityFieldsSupported
	return support
}

func validateEntityFieldsSupport(support EntityFieldsSupport) error {
	if support.State != EntityFieldsDeferred && support.State != EntityFieldsSupported {
		return entityFieldsError(EntityFieldsUnknownStateDiagnostic, "unknown EntityFields support state", SourceSpan{}, ErrEntityFieldsUnknownState)
	}
	profile := support.Profile
	if profile.ID == "" && profile.Version == 0 && profile.Digest == "" {
		return entityFieldsError(EntityFieldsUnboundProfileDiagnostic, "EntityFields profile is unbound", SourceSpan{}, ErrEntityFieldsUnboundProfile)
	}
	if profile.ID != syntax.EntityFieldsProfileID || profile.Version != syntax.EntityFieldsProfileVersion {
		return entityFieldsError(EntityFieldsProfileMismatchDiagnostic, "EntityFields profile identity or version does not match", SourceSpan{}, ErrEntityFieldsProfileMismatch)
	}
	if profile.Digest != syntax.EntityFieldsProfileDigest {
		return entityFieldsError(EntityFieldsProfileDigestDiagnostic, "EntityFields profile digest does not match", SourceSpan{}, ErrEntityFieldsProfileDigest)
	}
	return nil
}

func entityFieldsError(code, message string, span SourceSpan, cause error) error {
	return &EntityFieldsError{Code: code, Message: message, Span: span, Cause: cause}
}

func entityFieldsActivation(support EntityFieldsSupport, hasFields bool, span SourceSpan) error {
	if !hasFields {
		return nil
	}
	if err := validateEntityFieldsSupport(support); err != nil {
		return err
	}
	if support.State == EntityFieldsDeferred && hasFields {
		return entityFieldsError(EntityFieldsDeferredDiagnostic, "entity fields are deferred and unsupported by the public syntax", span, ErrEntityFieldsDeferred)
	}
	return nil
}

func entityFieldsProfileError(field Field, code, message string) error {
	return entityFieldsError(code, fmt.Sprintf("field %q: %s", field.ID, message), field.Span, nil)
}

func validateEntityFieldsModel(nodes []Node, registry semantic.TypeRegistry, support EntityFieldsSupport) error {
	if !modelHasFields(nodes) {
		return nil
	}
	if err := entityFieldsActivation(support, modelHasFields(nodes), firstModelFieldSpan(nodes)); err != nil {
		return err
	}
	if err := validateModelFields(nodes, registry); err != nil {
		return classifyEntityFieldsModelError(err, firstModelFieldSpan(nodes))
	}
	for _, node := range nodes {
		if err := validateEntityFieldSequence(node.ID, node.Fields); err != nil {
			return err
		}
		for _, field := range node.Fields {
			if err := validateSourceField(field, node.ID, registry); err != nil {
				return err
			}
			if err := validateEntityFieldsProfileField(field, registry); err != nil {
				return err
			}
		}
	}
	return nil
}

func classifyEntityFieldsModelError(err error, span SourceSpan) error {
	message := err.Error()
	switch {
	case strings.Contains(message, semantic.ErrAmbiguousType.Error()):
		return entityFieldsError(EntityFieldsAmbiguousTypeDiagnostic, message, span, semantic.ErrAmbiguousType)
	case strings.Contains(message, semantic.ErrUnknownType.Error()):
		return entityFieldsError(EntityFieldsUnknownTypeDiagnostic, message, span, semantic.ErrUnknownType)
	case strings.Contains(message, "parent is"):
		return entityFieldsError(EntityFieldsWrongParentDiagnostic, message, span, ErrInvalidField)
	default:
		return err
	}
}

func validateEntityFieldsProfileField(field Field, registry semantic.TypeRegistry) error {
	normalized, err := normalizeField(field, field.Parent, EntityKind, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	typeID, err := resolvedFieldTypeID(normalized, normalized.TypeRef, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	if typeID != semantic.BuiltinStringTypeID {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedTypeDiagnostic, fmt.Sprintf("resolved type %q is not in the bound profile", typeID))
	}
	if normalized.Presence != FieldPresenceRequired || normalized.Cardinality != FieldCardinalityOne {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedShapeDiagnostic, "only required × one is in the bound profile")
	}
	return nil
}

func classifyEntityFieldsTypeError(field Field, err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, semantic.ErrAmbiguousType.Error()):
		return entityFieldsProfileError(field, EntityFieldsAmbiguousTypeDiagnostic, message)
	case strings.Contains(message, semantic.ErrUnknownType.Error()):
		return entityFieldsProfileError(field, EntityFieldsUnknownTypeDiagnostic, message)
	default:
		return err
	}
}

func modelHasFields(nodes []Node) bool {
	for _, node := range nodes {
		if len(node.Fields) != 0 {
			return true
		}
	}
	return false
}

func firstModelFieldSpan(nodes []Node) SourceSpan {
	for _, node := range nodes {
		if len(node.Fields) != 0 {
			return node.Fields[0].Span
		}
	}
	return SourceSpan{}
}

func documentHasFields(document Document) bool {
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) != 0 {
			return true
		}
	}
	return false
}

func firstDocumentFieldSpan(document Document) SourceSpan {
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) != 0 {
			return declaration.Fields[0].Span
		}
	}
	return SourceSpan{}
}

func validateEntityFieldsDocument(document Document, namespace string, registry semantic.TypeRegistry, support EntityFieldsSupport) error {
	if err := entityFieldsActivation(support, documentHasFields(document), firstDocumentFieldSpan(document)); err != nil {
		return err
	}
	if !documentHasFields(document) {
		return nil
	}
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) == 0 {
			continue
		}
		if declaration.Kind != EntityKind {
			return entityFieldsError(EntityFieldsWrongParentDiagnostic, "fields are only valid on Entity declarations", declaration.Span, ErrInvalidField)
		}
		owner, err := declarationIdentity(namespace, declaration)
		if err != nil {
			return err
		}
		if err := validateEntityFieldSequence(owner, declaration.Fields); err != nil {
			return err
		}
		for _, field := range declaration.Fields {
			if err := validateSourceField(field, owner, registry); err != nil {
				return err
			}
			if err := validateEntityFieldsProfileField(field, registry); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEntityFieldSequence(owner ID, fields []Field) error {
	if len(fields) < 2 {
		return nil
	}
	file := fields[0].Span.File
	for index := 1; index < len(fields); index++ {
		if fields[index].Span.File != file || fields[index-1].Span.Start > fields[index].Span.Start {
			return entityFieldsError(EntityFieldsIllegalReorderDiagnostic, fmt.Sprintf("field order for entity %s is not source ordered", owner), fields[index].Span, ErrInvalidField)
		}
	}
	return nil
}
