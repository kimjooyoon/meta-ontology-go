package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
