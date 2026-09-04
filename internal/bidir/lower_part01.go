package bidir

import (
	"context"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ErrLowerCanceled is returned for cancellation and deadline expiration.
// It is intentionally stable so callers do not need to inspect context text.
var ErrLowerCanceled = errors.New("bidir lowering canceled")

// Lower parses the current syntax AST boundary into semantic IR.
func Lower(file *syntax.File) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}
func lowerWithEntityFieldsSupport(file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, semantic.DefaultTypeRegistry(), support)
}

// LowerWithTypes lowers the syntax carrier and resolves latent field TypeRefs
// through the supplied semantic registry.
func LowerWithTypes(file *syntax.File, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, registry, CurrentEntityFieldsSupport())
}
func lowerWithTypesAndEntityFieldsSupport(file *syntax.File, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(context.Background(), file, registry, support)
}

// LowerContext lowers without mutating file and never returns a partial IR.
func LowerContext(ctx context.Context, file *syntax.File) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}

// LowerContextWithTypes is the cancellable typed syntax lowerer.
func LowerContextWithTypes(ctx context.Context, file *syntax.File, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, registry, CurrentEntityFieldsSupport())
}
func lowerContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, semantic.DefaultTypeRegistry(), support)
}
func lowerContextWithTypesAndEntityFieldsSupport(ctx context.Context, file *syntax.File, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupportAndImplicitActivityPorts(ctx, file, registry, support, false)
}
func lowerContextWithTypesAndEntityFieldsSupportAndImplicitActivityPorts(ctx context.Context, file *syntax.File, registry semantic.TypeRegistry, support EntityFieldsSupport, allowImplicitActivityPorts bool) (semantic.IR, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return semantic.IR{}, err
	}
	document, err := documentFromSyntaxContextWithEntityFieldsSupport(ctx, file, support)
	if err != nil {
		return semantic.IR{}, err
	}
	document.ImplicitActivityPorts = allowImplicitActivityPorts
	ir, err := lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx, document, registry, support)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, err
	}
	return ir, nil
}
func nonNilLowerContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
func checkLowerContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrLowerCanceled
	default:
		return nil
	}
}

// DocumentFromSyntax adapts syntax without making the generic lens depend on
// parser implementation details.
func DocumentFromSyntax(file *syntax.File) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(context.Background(), file, CurrentEntityFieldsSupport())
}
