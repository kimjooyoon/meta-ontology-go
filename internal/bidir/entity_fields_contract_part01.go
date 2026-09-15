package bidir

import (
	"errors"
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
