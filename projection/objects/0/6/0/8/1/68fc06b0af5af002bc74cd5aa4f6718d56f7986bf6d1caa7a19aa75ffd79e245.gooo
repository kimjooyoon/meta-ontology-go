package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func documentFromSyntaxWithEntityFieldsSupport(file *syntax.File, support EntityFieldsSupport) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(context.Background(), file, support)
}

// DocumentFromSyntaxContext is the cancellable syntax adapter.
func DocumentFromSyntaxContext(ctx context.Context, file *syntax.File) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(ctx, file, CurrentEntityFieldsSupport())
}
func documentFromSyntaxContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (Document, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return Document{}, err
	}
	if file == nil || file.Package == nil {
		return Document{}, fmt.Errorf("package is required")
	}
	if file.Namespace == nil || file.Namespace.Name == "" {
		return Document{}, fmt.Errorf("namespace is required")
	}
	if err := entityFieldsActivation(support, syntaxFileHasFields(file), firstSyntaxFieldSpan(file)); err != nil {
		return Document{}, err
	}
	document := Document{Package: file.Package.Name, Namespace: file.Namespace.Name}
	for _, declaration := range syntaxDeclarations(file) {
		if err := checkLowerContext(ctx); err != nil {
			return Document{}, err
		}
		adapted, err := adaptSyntaxDeclaration(ctx, declaration)
		if err != nil {
			return Document{}, err
		}
		document.Declarations = append(document.Declarations, adapted)
	}
	if err := validateEntityFieldsDocument(document, document.Namespace, semantic.DefaultTypeRegistry(), support); err != nil {
		return Document{}, err
	}
	return document, nil
}
func syntaxFileHasFields(file *syntax.File) bool {
	if file == nil {
		return false
	}
	for _, declaration := range syntaxDeclarations(file) {
		entity, ok := declaration.(*syntax.EntityDecl)
		if ok && len(entity.Fields) > 0 {
			return true
		}
	}
	return false
}
func firstSyntaxFieldSpan(file *syntax.File) SourceSpan {
	if file == nil {
		return SourceSpan{}
	}
	for _, declaration := range syntaxDeclarations(file) {
		entity, ok := declaration.(*syntax.EntityDecl)
		if ok && len(entity.Fields) > 0 {
			return toSourceSpan(entity.Fields[0].Span)
		}
	}
	return SourceSpan{}
}
