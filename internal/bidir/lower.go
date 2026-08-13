package bidir

import (
	"context"
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ErrLowerCanceled is returned for cancellation and deadline expiration.
// It is intentionally stable so callers do not need to inspect context text.
var ErrLowerCanceled = errors.New("bidir lowering canceled")

// Lower parses the current syntax AST boundary into semantic IR.
func Lower(file *syntax.File) (semantic.IR, error) {
	return LowerContext(context.Background(), file)
}

// LowerContext lowers without mutating file and never returns a partial IR.
func LowerContext(ctx context.Context, file *syntax.File) (semantic.IR, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return semantic.IR{}, err
	}
	document, err := DocumentFromSyntaxContext(ctx, file)
	if err != nil {
		return semantic.IR{}, err
	}
	return LowerDocumentContext(ctx, document)
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
	return DocumentFromSyntaxContext(context.Background(), file)
}

// DocumentFromSyntaxContext is the cancellable syntax adapter.
func DocumentFromSyntaxContext(ctx context.Context, file *syntax.File) (Document, error) {
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
	return document, nil
}

func syntaxDeclarations(file *syntax.File) []syntax.Declaration {
	if len(file.Declarations) > 0 {
		return file.Declarations
	}
	return file.Decls
}

func adaptSyntaxDeclaration(ctx context.Context, declaration syntax.Declaration) (Declaration, error) {
	if err := checkLowerContext(ctx); err != nil {
		return Declaration{}, err
	}
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		return Declaration{Kind: EntityKind, ID: ID(value.ID), Name: value.Name, Span: toSourceSpan(value.Span)}, nil
	case *syntax.ActivityDecl:
		return adaptSyntaxActivity(ctx, value)
	default:
		return Declaration{}, fmt.Errorf("unsupported syntax declaration %T", declaration)
	}
}

func adaptSyntaxActivity(ctx context.Context, activity *syntax.ActivityDecl) (Declaration, error) {
	if err := checkLowerContext(ctx); err != nil {
		return Declaration{}, err
	}
	declaration := Declaration{Kind: ActivityKind, Name: activity.Name, Span: toSourceSpan(activity.Span)}
	if len(activity.Inputs) == 0 && len(activity.Parameters) != 0 {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Parameters; canonical Inputs is required", activity.Name)
	}
	for _, input := range activity.Inputs {
		if err := checkLowerContext(ctx); err != nil {
			return Declaration{}, err
		}
		declaration.Inputs = append(declaration.Inputs, Reference{Name: input.Name, Span: toSourceSpan(input.Span)})
	}
	if activity.Output == "" && activity.Result.Name != "" {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Result; canonical Output is required", activity.Name)
	}
	if activity.Output != "" {
		declaration.Outputs = append(declaration.Outputs, Reference{Name: activity.Output, Span: toSourceSpan(activity.Span)})
	}
	return declaration, nil
}

func toSourceSpan(span syntax.Span) SourceSpan {
	return SourceSpan{
		File:        span.Filename,
		Start:       span.Start.Offset,
		End:         span.End.Offset,
		StartLine:   span.Start.Line,
		StartColumn: span.Start.Column,
		EndLine:     span.End.Line,
		EndColumn:   span.End.Column,
	}
}

func toSemanticSpan(span SourceSpan) semantic.Span {
	return semantic.Span{
		File:  span.File,
		Start: semantic.Position{Offset: span.Start, Line: span.StartLine, Column: span.StartColumn},
		End:   semantic.Position{Offset: span.End, Line: span.EndLine, Column: span.EndColumn},
	}
}
