package cycles

import (
	"fmt"
	"strings"
)

func edgeIdentityKey(input edgeInput) string {
	subject := input.subject
	if subject == "" {
		subject = "invalid:" + strings.TrimSpace(input.edge.Subject)
	}
	object := input.object
	if object == "" {
		object = "invalid:" + strings.TrimSpace(input.edge.Object)
	}
	return strings.Join([]string{subject, string(input.predicate), object}, "\x00")
}
func unresolvedDiagnostic(edge Edge, endpoint string, canonical ID) Diagnostic {
	raw := edge.Subject
	if endpoint == "object" {
		raw = edge.Object
	}
	if canonical == "" {
		return Diagnostic{
			Code: InvalidStableID, Subject: strings.TrimSpace(edge.Subject),
			Predicate: relationName(edge), Object: strings.TrimSpace(edge.Object), Span: edge.Span,
			Message: fmt.Sprintf("%s stable ID %q is invalid", endpoint, raw),
		}
	}
	return Diagnostic{
		Code: UnresolvedStableID, Subject: strings.TrimSpace(edge.Subject),
		Predicate: relationName(edge), Object: strings.TrimSpace(edge.Object), Span: edge.Span,
		Message: fmt.Sprintf("%s stable ID %q is not declared", endpoint, canonical),
	}
}
func validDirection(predicate Relation, subject, object Kind) bool {
	expectedSubject, expectedObject, known := expectedKinds(predicate)
	return known && subject == expectedSubject && object == expectedObject
}
