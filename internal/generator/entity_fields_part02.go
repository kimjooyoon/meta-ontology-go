package generator

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func validateEntityFieldsSupport(support syntax.EntityFieldsSupport) error {
	switch support.State {
	case syntax.EntityFieldsDeferred, syntax.EntityFieldsSupported:
	default:
		return entityFieldsError(entityFieldsUnknownStateDiagnostic, Field{}, fmt.Sprintf("unknown support state %q", support.State))
	}
	if support.Profile.ID == "" && support.Profile.Version == 0 && support.Profile.Digest == "" {
		return entityFieldsError(entityFieldsUnboundProfileDiagnostic, Field{}, "profile is unbound")
	}
	if support.Profile.ID != syntax.EntityFieldsProfileID && support.Profile.ID != syntax.EntityFieldsV2ProfileID && support.Profile.ID != syntax.EntityFieldsV3ProfileID ||
		support.Profile.ID == syntax.EntityFieldsProfileID && support.Profile.Version != syntax.EntityFieldsProfileVersion ||
		support.Profile.ID == syntax.EntityFieldsV2ProfileID && support.Profile.Version != syntax.EntityFieldsV2ProfileVersion ||
		support.Profile.ID == syntax.EntityFieldsV3ProfileID && support.Profile.Version != syntax.EntityFieldsV3ProfileVersion {
		return entityFieldsError(entityFieldsProfileMismatchDiagnostic, Field{}, "profile identity or version does not match")
	}
	if err := support.Profile.Validate(); err != nil {
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
	return validateSupportedEntityFields(ir, support.Profile.ID)
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
