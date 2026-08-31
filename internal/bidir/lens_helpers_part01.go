package bidir

import (
	"fmt"
	"strings"
)

func declarationIdentity(namespace string, declaration Declaration) (ID, error) {
	if declaration.ID != "" {
		if err := validateID(declaration.ID); err != nil {
			return "", fmt.Errorf("declaration %q: %w", declaration.Name, err)
		}
		return declaration.ID, nil
	}
	derived := slug(declaration.Name)
	if derived == "" {
		return "", fmt.Errorf("declaration %q cannot derive a stable ID", declaration.Name)
	}
	return ID(strings.ReplaceAll(namespace, "_", "-") + "://" + strings.ToLower(string(declaration.Kind)) + "/" + derived), nil
}
func implicitEntityDeclarations(declarations []Declaration, namespace string) []Declaration {
	seen := make(map[string]struct{}, len(declarations))
	for _, declaration := range declarations {
		seen[referenceKey(namespace, declaration.Name)] = struct{}{}
	}
	result := make([]Declaration, 0)
	for _, declaration := range declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		references := append(append([]Reference(nil), declaration.Inputs...), declaration.Outputs...)
		for _, reference := range references {
			refNamespace := reference.Namespace
			if refNamespace == "" {
				refNamespace = namespace
			}
			if refNamespace != namespace || strings.TrimSpace(reference.Name) == "" {
				continue
			}
			key := referenceKey(refNamespace, reference.Name)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, Declaration{Kind: EntityKind, Name: reference.Name, Span: reference.Span})
		}
	}
	return result
}
func validateDocumentSpans(document Document) error {
	for index, declaration := range document.Declarations {
		if err := declaration.Span.Validate(); err != nil {
			return fmt.Errorf("declaration %d %q: %w", index, declaration.Name, err)
		}
		for fieldIndex, field := range declaration.Fields {
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
					return fmt.Errorf("declaration %q field %d %s: %w", declaration.Name, fieldIndex, item.name, err)
				}
			}
		}
		for refIndex, reference := range declaration.Inputs {
			if err := reference.Span.Validate(); err != nil {
				return fmt.Errorf("declaration %q input %d: %w", declaration.Name, refIndex, err)
			}
		}
		for refIndex, reference := range declaration.Outputs {
			if err := reference.Span.Validate(); err != nil {
				return fmt.Errorf("declaration %q output %d: %w", declaration.Name, refIndex, err)
			}
		}
	}
	for index, relation := range document.Relations {
		if err := relation.Span.Validate(); err != nil {
			return fmt.Errorf("relation %d %s %q -> %q: %w", index, relation.Kind, relation.Source, relation.Target, err)
		}
	}
	return nil
}
