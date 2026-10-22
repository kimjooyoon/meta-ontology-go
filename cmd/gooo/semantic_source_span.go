package main

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func semanticFailureSpan(file *syntax.File, err error) syntax.Span {
	var reference *bidir.ReferenceResolutionError
	if errors.As(err, &reference) && reference.Span.Valid() {
		return syntax.Span{
			Filename: reference.Span.File,
			Start:    syntax.Position{Offset: reference.Span.Start, Line: reference.Span.StartLine, Column: reference.Span.StartColumn},
			End:      syntax.Position{Offset: reference.Span.End, Line: reference.Span.EndLine, Column: reference.Span.EndColumn},
		}
	}
	return syntaxFileSpan(file)
}
