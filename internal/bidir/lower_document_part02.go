package bidir

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx context.Context, document Document, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return semantic.IR{}, err
	}
	if err := validateDocumentSpans(document); err != nil {
		return semantic.IR{}, err
	}
	namespaceText := strings.TrimSpace(document.Namespace)
	if namespaceText == "" {
		namespaceText = "gooo"
	}
	namespace, err := semantic.ParseNamespace(namespaceText)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := validateEntityFieldsDocument(document, namespace.String(), registry, support); err != nil {
		return semantic.IR{}, err
	}
	ir := semantic.NewIR(document.Package, namespace)
	names, ids, err := lowerDocumentNodes(ctx, &ir, document, namespace)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := lowerDocumentContracts(ctx, &ir, document, namespace, ids, names); err != nil {
		return semantic.IR{}, err
	}
	if err := lowerDocumentRelations(ctx, &ir, document.Relations); err != nil {
		return semantic.IR{}, err
	}
	if err := lowerDocumentRuntimeBindings(ctx, &ir, document, namespace, ids, names); err != nil {
		return semantic.IR{}, err
	}
	for _, policy := range document.Policies {
		if err := checkLowerContext(ctx); err != nil {
			return semantic.IR{}, err
		}
		normalized, err := policy.Normalized()
		if err != nil {
			return semantic.IR{}, err
		}
		ir.Policies = append(ir.Policies, normalized)
	}
	if err := validateLoweredContext(ctx, ir); err != nil {
		return semantic.IR{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, err
	}
	if semanticIRHasFields(ir) {
		normalized, err := ir.NormalizedWithTypes(registry)
		if err != nil {
			return semantic.IR{}, err
		}
		ir = normalized
	}
	if err := checkLowerContext(ctx); err != nil {
		return semantic.IR{}, err
	}
	return ir, nil
}
