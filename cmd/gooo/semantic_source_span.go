package main

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func semanticFailureSpan(file *syntax.File, err error) syntax.Span {
	var reference *bidir.ReferenceResolutionError
	var field *bidir.EntityFieldsError
	var span bidir.SourceSpan
	switch {
	case errors.As(err, &reference) && reference.Span.Valid():
		span = reference.Span
	case errors.As(err, &field) && field.Span.Valid():
		span = field.Span
	default:
		return syntaxFileSpan(file)
	}
	return syntax.Span{
		Filename: span.File,
		Start:    syntax.Position{Offset: span.Start, Line: span.StartLine, Column: span.StartColumn},
		End:      syntax.Position{Offset: span.End, Line: span.EndLine, Column: span.EndColumn},
	}
}
