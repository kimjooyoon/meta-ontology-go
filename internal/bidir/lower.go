package bidir

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Lower parses the current syntax AST boundary into semantic IR.
func Lower(file *syntax.File) (semantic.IR, error) {
	document, err := DocumentFromSyntax(file)
	if err != nil {
		return semantic.IR{}, err
	}
	return LowerDocument(document)
}

// DocumentFromSyntax adapts syntax without making the generic lens depend on
// parser implementation details.
func DocumentFromSyntax(file *syntax.File) (Document, error) {
	if file == nil || file.Package == nil {
		return Document{}, fmt.Errorf("package is required")
	}
	if file.Namespace == nil || file.Namespace.Name == "" {
		return Document{}, fmt.Errorf("namespace is required")
	}
	document := Document{Package: file.Package.Name, Namespace: file.Namespace.Name}
	for _, declaration := range syntaxDeclarations(file) {
		adapted, err := adaptSyntaxDeclaration(declaration)
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

func adaptSyntaxDeclaration(declaration syntax.Declaration) (Declaration, error) {
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		return Declaration{Kind: EntityKind, ID: ID(value.ID), Name: value.Name, Span: toSourceSpan(value.Span)}, nil
	case *syntax.ActivityDecl:
		return adaptSyntaxActivity(value)
	default:
		return Declaration{}, fmt.Errorf("unsupported syntax declaration %T", declaration)
	}
}

func adaptSyntaxActivity(activity *syntax.ActivityDecl) (Declaration, error) {
	declaration := Declaration{Kind: ActivityKind, Name: activity.Name, Span: toSourceSpan(activity.Span)}
	if len(activity.Inputs) == 0 && len(activity.Parameters) != 0 {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Parameters; canonical Inputs is required", activity.Name)
	}
	for _, input := range activity.Inputs {
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
