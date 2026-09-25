package bidir

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// LowerDocument lowers the parser-neutral view into the current semantic IR.
func LowerDocument(document Document) (semantic.IR, error) {
	return lowerDocumentWithTypesAndEntityFieldsSupport(document, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}
func lowerDocumentWithEntityFieldsSupport(document Document, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerDocumentWithTypesAndEntityFieldsSupport(document, semantic.DefaultTypeRegistry(), support)
}

// LowerDocumentWithTypes lowers a parser-neutral document and resolves latent
// field TypeRefs through the supplied semantic registry.
func LowerDocumentWithTypes(document Document, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerDocumentWithTypesAndEntityFieldsSupport(document, registry, CurrentEntityFieldsSupport())
}
func lowerDocumentWithTypesAndEntityFieldsSupport(document Document, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerDocumentContextWithTypesAndEntityFieldsSupport(context.Background(), document, registry, support)
}

// LowerDocumentContext is the cancellable parser-neutral lowerer.
func LowerDocumentContext(ctx context.Context, document Document) (semantic.IR, error) {
	return lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx, document, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}

// LowerDocumentContextWithTypes is the cancellable typed parser-neutral
// lowerer. It returns no partial IR on any field or registry failure.
func LowerDocumentContextWithTypes(ctx context.Context, document Document, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx, document, registry, CurrentEntityFieldsSupport())
}
func lowerDocumentContextWithEntityFieldsSupport(ctx context.Context, document Document, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx, document, semantic.DefaultTypeRegistry(), support)
}
