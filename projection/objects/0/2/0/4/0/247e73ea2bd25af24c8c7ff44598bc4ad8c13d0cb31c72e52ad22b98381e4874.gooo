package generator

import (
	"fmt"
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
